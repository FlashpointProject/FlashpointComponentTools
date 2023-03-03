using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using System.Windows.Forms;

using Downloader;
using FlashpointInstaller.Common;
using Microsoft.WindowsAPICodePack.Taskbar;
using SharpCompress.Archives.Zip;
using SharpCompress.Common;
using SharpCompress.Readers;

namespace FlashpointInstaller
{
    public partial class Operate : Form
    {
        Component workingComponent;

        List<Component> markedComponents   = new List<Component>();

        DownloadService downloader = new DownloadService();

        Stream stream;
        ZipArchive archive;
        IReader reader;

        long byteProgress = 0;
        long byteTotal = 0;

        int cancelStatus = 0;

        public Operate() => InitializeComponent();

        private async void Operation_Load(object sender, EventArgs e)
        {
            TaskbarManager.Instance.SetProgressState(TaskbarProgressBarState.Normal, FPM.Main.Handle);

            downloader.DownloadProgressChanged += OnDownloadProgressChanged;
            downloader.DownloadFileCompleted += OnDownloadFileCompleted;

            FPM.IterateList(FPM.Main.ComponentList.Nodes, node =>
            {
                if (node.Checked && node.Tag.GetType().ToString().EndsWith("Component"))
                {
                    markedComponents.Add(node.Tag as Component);
                }
            });

            byteTotal = FPM.SizeTracker.ToDownload;

            foreach (var component in markedComponents)
            {
                workingComponent = component;
                if (component.Size > 0) stream = await downloader.DownloadFileTaskAsync(component.URL);

                if (cancelStatus != 0) return;

                await Task.Run(ExtractComponents);

                byteProgress += component.Size;
            }

            if (cancelStatus == 0)
            {
                if (!FPM.RedistInstalled)
                {
                    ProgressLabel.Text = "Installing Visual C++ 2015 x86 redistributable...";
                    CancelButton.Enabled = false;

                    string redistPath = Path.GetTempPath() + "vc_redist.x86.exe";

                    if (!File.Exists(redistPath))
                    {
                        await new DownloadService().DownloadFileTaskAsync("https://aka.ms/vs/17/release/vc_redist.x86.exe", redistPath);
                    }

                    await Task.Run(() =>
                    {
                        var redistProcess = Process.Start(redistPath, "/install /norestart /quiet");
                        redistProcess.WaitForExit();

                        if (redistProcess.ExitCode != 0)
                        {
                            MessageBox.Show(
                                "Failed to install Visual C++ 2015 x86 redistributable.\n\n" +
                                "You can try installing it manually from https://aka.ms/vs/17/release/vc_redist.x86.exe.",
                                "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                            );
                        }
                    });
                }

                FinishOperation();
            }
        }

        private void OnDownloadProgressChanged(object sender, DownloadProgressChangedEventArgs e)
        {
            if (cancelStatus != 0)
            {
                downloader.CancelAsync();
                return;
            }

            double currentProgress = (double)e.ReceivedBytesSize / e.TotalBytesToReceive;
            long currentSize = workingComponent.Size;
            double totalProgress = (byteProgress + (currentProgress / 2 * currentSize)) / byteTotal;

            ProgressMeasure.Invoke((MethodInvoker)delegate
            {
                ProgressMeasure.Value = (int)((double)totalProgress * ProgressMeasure.Maximum);
            });

            ProgressLabel.Invoke((MethodInvoker)delegate
            {
                ProgressLabel.Text =
                    $"[{(int)((double)totalProgress * 100)}%] Downloading component \"{workingComponent.Title}\"... " +
                    $"{FPM.GetFormattedBytes(e.ReceivedBytesSize)} of {FPM.GetFormattedBytes(e.TotalBytesToReceive)}";
            });

            FPM.Main.Invoke((MethodInvoker)delegate
            {
                TaskbarManager.Instance.SetProgressValue(
                    (int)((double)totalProgress * ProgressMeasure.Maximum), ProgressMeasure.Maximum, FPM.Main.Handle
                );
            });
        }

        private void OnDownloadFileCompleted(object sender, System.ComponentModel.AsyncCompletedEventArgs e)
        {
            if (e.Cancelled) cancelStatus = 2;
        }

        private void ExtractComponents()
        {
            string rootPath = FPM.DestinationPath;
            string infoPath = Path.Combine(rootPath, "Components");
            string infoFile = Path.Combine(infoPath, $"{workingComponent.ID}.txt");

            Directory.CreateDirectory(infoPath);

            using (TextWriter writer = File.CreateText(infoFile))
            {
                string[] header = new List<string>
                    { workingComponent.Hash, $"{workingComponent.Size}" }.Concat(workingComponent.Depends).ToArray();

                writer.WriteLine(string.Join(" ", header));
            }

            if (workingComponent.Size == 0) return;

            using (archive = ZipArchive.Open(stream))
            {
                using (reader = archive.ExtractAllEntries())
                {
                    long extractedSize = 0;
                    long totalSize = archive.TotalUncompressSize;
                    
                    string destPath = Path.Combine(rootPath, workingComponent.Path.Replace('/', '\\'));

                    while (cancelStatus == 0 && reader.MoveToNextEntry())
                    {
                        if (reader.Entry.IsDirectory) continue;

                        Directory.CreateDirectory(destPath);

                        reader.WriteEntryToDirectory(destPath, new ExtractionOptions {
                            ExtractFullPath = true, Overwrite = true, PreserveFileTime = true 
                        });

                        using (TextWriter writer = File.AppendText(infoFile))
                        {
                            writer.WriteLine(Path.Combine(workingComponent.Path, reader.Entry.Key).Replace("/", @"\"));
                        }

                        extractedSize += reader.Entry.Size;

                        double currentProgress = (double)extractedSize / totalSize;
                        long currentSize = workingComponent.Size;
                        double totalProgress = (byteProgress + (currentSize / 2) + (currentProgress / 2 * currentSize)) / byteTotal;

                        ProgressMeasure.Invoke((MethodInvoker)delegate
                        {
                            ProgressMeasure.Value = (int)((double)totalProgress * ProgressMeasure.Maximum);
                        });

                        ProgressLabel.Invoke((MethodInvoker)delegate
                        {
                            ProgressLabel.Text = 
                                $"[{(int)((double)totalProgress * 100)}%] Extracting component \"{workingComponent.Title}\"... " +
                                $"{FPM.GetFormattedBytes(extractedSize)} of {FPM.GetFormattedBytes(totalSize)}";
                        });

                        FPM.Main.Invoke((MethodInvoker)delegate
                        {
                            TaskbarManager.Instance.SetProgressValue(
                                (int)((double)totalProgress * ProgressMeasure.Maximum), ProgressMeasure.Maximum, FPM.Main.Handle
                            );
                        });
                    }

                    if (cancelStatus != 0)
                    {
                        reader.Cancel();
                        cancelStatus = 2;
                    }
                }
            }
        }

        private async void FinishOperation()
        {
            TaskbarManager.Instance.SetProgressState(TaskbarProgressBarState.NoProgress, FPM.Main.Handle);

            await Task.Run(() =>
            {
                var shortcutPaths = new List<string>();

                if (FPM.Main.ShortcutDesktop.Checked)
                {
                    shortcutPaths.Add(Environment.GetFolderPath(Environment.SpecialFolder.Desktop));
                }
                if (FPM.Main.ShortcutStartMenu.Checked)
                {
                    shortcutPaths.Add(Environment.GetFolderPath(Environment.SpecialFolder.StartMenu));
                }

                foreach (string path in shortcutPaths)
                {
                    var shortcut = new IWshRuntimeLibrary.WshShell().CreateShortcut(Path.Combine(path, "Flashpoint.lnk"));
                    shortcut.TargetPath = Path.Combine(FPM.DestinationPath, "Launcher", "Flashpoint.exe");
                    shortcut.WorkingDirectory = Path.Combine(FPM.DestinationPath, "Launcher");
                    shortcut.Description = "Shortcut to Flashpoint";
                    shortcut.Save();
                }
            });

            Hide();
            FPM.Main.Hide();

            var finishWindow = new Finish();
            finishWindow.ShowDialog();
        }

        private async void CancelButton_Click(object sender, EventArgs e)
        {
            cancelStatus = 1;

            CancelButton.Enabled = false;
            ProgressLabel.Invoke((MethodInvoker)delegate
            {
                ProgressLabel.Text = "Cancelling...";
            });

            await Task.Run(() =>
            {
                while (cancelStatus != 2) { }

                if (Directory.Exists(FPM.DestinationPath))
                {
                    foreach (string file in Directory.EnumerateFiles(FPM.DestinationPath))
                    {
                        try { File.Delete(file); } catch { }
                    }

                    try { Directory.Delete(FPM.DestinationPath, true); } catch { }
                }
            });

            TaskbarManager.Instance.SetProgressState(TaskbarProgressBarState.NoProgress, FPM.Main.Handle);

            Close();
        }

        private void Operation_FormClosing(object sender, FormClosingEventArgs e)
        {
            if (cancelStatus != 2) e.Cancel = true;
        }
    }
}