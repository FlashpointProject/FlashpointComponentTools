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
    public partial class Operation : Form
    {
        Component workingComponent;

        List<Component> addedComponents   = new List<Component>();
        List<Component> removedComponents = new List<Component>();

        DownloadService downloader = new DownloadService();

        Stream stream;
        ZipArchive archive;
        IReader reader;

        long byteProgress = 0;
        long byteTotal = 0;

        int cancelStatus = 0;

        public Operation() => InitializeComponent();

        private async void Operation_Load(object sender, EventArgs e)
        {
            TaskbarManager.Instance.SetProgressState(TaskbarProgressBarState.Normal, FPM.Main.Handle);

            downloader.DownloadProgressChanged += OnDownloadProgressChanged;
            downloader.DownloadFileCompleted += OnDownloadFileCompleted;

            if (FPM.OperateMode == 0)
            {
                FPM.Iterate(FPM.Main.ComponentList.Nodes, node =>
                {
                    if (node.Checked && node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        addedComponents.Add(node.Tag as Component);
                    }
                });

                byteTotal = FPM.SizeTracker.ToDownload;
            }
            else if (FPM.OperateMode != 3)
            {
                Text = "Modifying Flashpoint...";
                CancelButton.Visible = false;

                if (FPM.OperateMode == 1)
                {
                    FPM.Iterate(FPM.Main.ComponentList2.Nodes, node =>
                    {
                        if (node.Tag.GetType().ToString().EndsWith("Component"))
                        {
                            var component = node.Tag as Component;

                            if (!node.Checked && FPM.ComponentTracker.Downloaded.Contains(component))
                            {
                                removedComponents.Add(component);
                            }
                            if (node.Checked && !FPM.ComponentTracker.Downloaded.Contains(component))
                            {
                                addedComponents.Add(component);
                            }
                        }
                    });
                }
                if (FPM.OperateMode == 2)
                {
                    foreach (var component in FPM.ComponentTracker.ToUpdate)
                    {
                        if (FPM.ComponentTracker.Downloaded.Contains(component))
                        {
                            removedComponents.Add(component);
                        }

                        addedComponents.Add(component);
                    }
                }

                byteTotal = removedComponents.Concat(addedComponents).Sum(item => item.Size);
            }
            else if (FPM.OperateMode == 3)
            {
                Text = "Removing Flashpoint...";
                CancelButton.Visible = false;

                await Task.Run(RemoveFlashpoint);

                FinishOperation();
            }

            foreach (var component in removedComponents)
            {
                workingComponent = component;

                await Task.Run(RemoveComponents);

                byteProgress += component.Size;
            }

            foreach (var component in addedComponents)
            {
                workingComponent = component;
                stream = await downloader.DownloadFileTaskAsync(component.URL);

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
            using (archive = ZipArchive.Open(stream))
            {
                using (reader = archive.ExtractAllEntries())
                {
                    long extractedSize = 0;
                    long totalSize = archive.TotalUncompressSize;

                    string rootPath = FPM.OperateMode > 0 ? FPM.SourcePath : FPM.DestinationPath;
                    string destPath = Path.Combine(rootPath, workingComponent.Path.Replace('/', '\\'));
                    string infoPath = Path.Combine(rootPath, "Components");
                    string infoFile = Path.Combine(infoPath, $"{workingComponent.ID}.txt");

                    Directory.CreateDirectory(infoPath);

                    using (TextWriter writer = File.CreateText(infoFile))
                    {
                        string[] header = new[] { workingComponent.Hash, workingComponent.Size.ToString() }.ToArray();

                        writer.WriteLine(string.Join(" ", header));
                    }

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

        private void RemoveComponents()
        {
            string infoFile = Path.Combine(FPM.SourcePath, "Components", $"{workingComponent.ID}.txt");
            string[] infoText = File.ReadAllLines(infoFile);

            long removedFiles = 0;
            long totalFiles = infoText.Length - 1;
            long totalSize = workingComponent.Size;

            for (int i = 1; i < infoText.Length; i++)
            {
                string filePath = Path.Combine(FPM.SourcePath, infoText[i]);
                double removeProgress = removedFiles / totalFiles;
                double totalProgress = (byteProgress + (removeProgress * totalSize)) / byteTotal;

                ProgressMeasure.Invoke((MethodInvoker)delegate
                {
                    ProgressMeasure.Value = (int)((double)totalProgress * ProgressMeasure.Maximum);
                });

                ProgressLabel.Invoke((MethodInvoker)delegate
                {
                    ProgressLabel.Text =
                        $"[{(int)((double)totalProgress * 100)}%] Removing component \"{workingComponent.Title}\"... " +
                        $"{removedFiles} of {totalFiles} files";
                });

                FPM.Main.Invoke((MethodInvoker)delegate
                {
                    TaskbarManager.Instance.SetProgressValue(
                        (int)((double)totalProgress * ProgressMeasure.Maximum), ProgressMeasure.Maximum, FPM.Main.Handle
                    );
                });

                FPM.DeleteFileAndDirectories(filePath);

                removedFiles++;
            }

            FPM.DeleteFileAndDirectories(infoFile);
        }

        private void RemoveFlashpoint()
        {
            string[] files = Directory.GetFiles(FPM.SourcePath2);

            for (int i = 0; i < files.Length; i++)
            {
                double progress = (i + 1) / files.Length;
                string fileName = files[i].Substring(files[i].LastIndexOf(@"\"));

                ProgressMeasure.Invoke((MethodInvoker)delegate
                {
                    ProgressMeasure.Value = (int)((double)progress * ProgressMeasure.Maximum);
                });

                ProgressLabel.Invoke((MethodInvoker)delegate
                {
                    ProgressLabel.Text = $"[{(int)((double)progress * 100)}%] Removing \"{fileName}\"... {i + 1} of {files.Length} files";
                });

                File.Delete(files[i]);
            }

            ProgressLabel.Invoke((MethodInvoker)delegate
            {
                ProgressLabel.Text = $"[100%] Removing directories...";
            });

            Directory.Delete(FPM.SourcePath2, true);
        }

        private async void FinishOperation()
        {
            if (FPM.OperateMode == 0)
            {
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

                var finishWindow = new FinishOperation();
                finishWindow.ShowDialog();
            }
            else if (FPM.OperateMode == 3)
            {
                if (FPM.Main.RemoveShortcuts.Checked)
                {
                    var shortcutPaths = new List<string>()
                    {
                        Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.StartMenu), "Flashpoint.lnk"),
                        Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.Desktop),   "Flashpoint.lnk")
                    };

                    foreach (string path in shortcutPaths) if (File.Exists(path)) File.Delete(path);
                }

                Hide();
                FPM.Main.Hide();

                var finishWindow = new FinishOperation();
                finishWindow.ShowDialog();
            }

            Close();
        }

        private async void CancelButton_Click(object sender, EventArgs e)
        {
            CancelButton.Enabled = false;
            cancelStatus = 1;

            await Task.Run(() =>
            { 
                while (cancelStatus != 2) { }

                if (Directory.Exists(FPM.DestinationPath)) Directory.Delete(FPM.DestinationPath, true);
            });

            TaskbarManager.Instance.SetProgressState(TaskbarProgressBarState.NoProgress, FPM.Main.Handle);
            
            Close();
        }
    }
}