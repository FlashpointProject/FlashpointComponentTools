using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using System.Windows.Forms;

using Downloader;
using FlashpointInstaller.Common;
using IWshRuntimeLibrary;
using SharpCompress.Archives.Zip;
using SharpCompress.Common;
using SharpCompress.Readers;

using File = System.IO.File;

namespace FlashpointInstaller
{
    public partial class Operation : Form
    {
        Dictionary<string, string> workingComponent;

        DownloadService downloader = new DownloadService(new DownloadConfiguration { OnTheFlyDownload = false });

        Stream stream;
        ZipArchive archive;
        IReader reader;

        long byteProgress = 0;
        long byteTotal;

        int cancelStatus = 0;

        public Operation() => InitializeComponent();

        private async void Operation_Load(object sender, EventArgs e)
        {
            if (FPM.DownloadMode != 0)
            {
                Text = "Modifying Flashpoint...";
                CancelButton.Visible = false;
            }

            downloader.DownloadProgressChanged += OnDownloadProgressChanged;
            downloader.DownloadFileCompleted += OnDownloadFileCompleted;

            if (FPM.DownloadMode != 0)
            {
                FPM.ComponentTracker.ToAdd.Clear();
                FPM.ComponentTracker.ToRemove.Clear();

                if (FPM.DownloadMode == 1)
                {
                    FPM.Iterate(FPM.Main.ComponentList2.Nodes, node =>
                    {
                        var attributes = node.Tag as Dictionary<string, string>;

                        if (attributes["type"] == "component")
                        {
                            if (!node.Checked && FPM.ComponentTracker.Downloaded.Contains(attributes))
                            {
                                FPM.ComponentTracker.ToRemove.Add(attributes);
                            }
                            if (node.Checked && !FPM.ComponentTracker.Downloaded.Contains(attributes))
                            {
                                FPM.ComponentTracker.ToAdd.Add(attributes);
                            }
                        }
                    });
                }
                else if (FPM.DownloadMode == 2)
                {
                    foreach (var component in FPM.ComponentTracker.ToUpdate)
                    {
                        FPM.ComponentTracker.ToRemove.Add(component);
                        FPM.ComponentTracker.ToAdd.Add(component);
                    }
                }

                byteTotal = 0;

                foreach (var component in FPM.ComponentTracker.ToRemove.Concat(FPM.ComponentTracker.ToAdd))
                {
                    byteTotal += long.Parse(component["size"]);
                }

                foreach (var component in FPM.ComponentTracker.ToRemove)
                {
                    workingComponent = component;

                    await Task.Run(RemoveTask);
                    
                    byteProgress += long.Parse(component["size"]);
                }

                foreach (var component in FPM.ComponentTracker.ToAdd)
                {
                    workingComponent = component;
                    stream = await downloader.DownloadFileTaskAsync(component["url"]);

                    await Task.Run(ExtractTask);

                    byteProgress += long.Parse(component["size"]);
                }
            }
            else
            {
                FPM.ComponentTracker.ToDownload.Clear();
                byteTotal = FPM.SizeTracker.ToDownload;

                FPM.Iterate(FPM.Main.ComponentList.Nodes, node =>
                {
                    var attributes = node.Tag as Dictionary<string, string>;

                    if (attributes["type"] == "component" && node.Checked) FPM.ComponentTracker.ToDownload.Add(attributes);
                });

                foreach (var component in FPM.ComponentTracker.ToDownload)
                {
                    workingComponent = component;
                    stream = await downloader.DownloadFileTaskAsync(component["url"]);

                    if (cancelStatus != 0) return;

                    Directory.CreateDirectory(FPM.DestinationPath);
                    await Task.Run(ExtractTask);

                    byteProgress += long.Parse(component["size"]);
                }
            }

            if (cancelStatus == 0) FinishDownload();
        }

        private void OnDownloadProgressChanged(object sender, DownloadProgressChangedEventArgs e)
        {
            if (cancelStatus != 0)
            {
                downloader.CancelAsync();
                return;
            }

            double currentProgress = (double)e.ReceivedBytesSize / e.TotalBytesToReceive;
            long currentSize = long.Parse(workingComponent["size"]);
            double totalProgress = (byteProgress + (currentProgress / 2 * currentSize)) / byteTotal;

            ProgressMeasure.Invoke((MethodInvoker)delegate
            {
                ProgressMeasure.Value = (int)((double)totalProgress * ProgressMeasure.Maximum);
            });

            ProgressLabel.Invoke((MethodInvoker)delegate
            {
                ProgressLabel.Text =
                    $"[{(int)((double)totalProgress * 100)}%] Downloading component \"{workingComponent["title"]}\"... " +
                    $"{FPM.GetFormattedBytes(e.ReceivedBytesSize)} of {FPM.GetFormattedBytes(e.TotalBytesToReceive)}";
            });
        }

        private void OnDownloadFileCompleted(object sender, AsyncCompletedEventArgs e)
        {
            if (e.Cancelled) cancelStatus = 2;
        }

        private void ExtractTask()
        {
            using (archive = ZipArchive.Open(stream))
            {
                using (reader = archive.ExtractAllEntries())
                {
                    long extractedSize = 0;
                    long totalSize = archive.TotalUncompressSize;

                    string destPath = FPM.DownloadMode > 0 ? FPM.SourcePath : FPM.DestinationPath;
                    string infoPath = Path.Combine(destPath, "Components", workingComponent["path"]);
                    string infoFile = Path.Combine(infoPath, $"{workingComponent["title"]}.txt");

                    Directory.CreateDirectory(infoPath);

                    using (TextWriter writer = File.CreateText(infoFile))
                    {
                        writer.WriteLine($"{workingComponent["hash"]} {workingComponent["size"]} {workingComponent["url"]}");
                    }

                    while (cancelStatus == 0 && reader.MoveToNextEntry())
                    {
                        if (reader.Entry.IsDirectory) continue;

                        reader.WriteEntryToDirectory(destPath, new ExtractionOptions {
                            ExtractFullPath = true, Overwrite = true, PreserveFileTime = true
                        });

                        using (TextWriter writer = File.AppendText(infoFile))
                        {
                            writer.WriteLine(reader.Entry.Key.Replace("/", @"\"));
                        }

                        extractedSize += reader.Entry.Size;

                        double currentProgress = (double)extractedSize / totalSize;
                        long currentSize = long.Parse(workingComponent["size"]);
                        double totalProgress = (byteProgress + (currentSize / 2) + (currentProgress / 2 * currentSize)) / byteTotal;

                        ProgressMeasure.Invoke((MethodInvoker)delegate
                        {
                            ProgressMeasure.Value = (int)((double)totalProgress * ProgressMeasure.Maximum);
                        });

                        ProgressLabel.Invoke((MethodInvoker)delegate
                        {
                            ProgressLabel.Text = 
                                $"[{(int)((double)totalProgress * 100)}%] Extracting component \"{workingComponent["title"]}\"... " +
                                $"{FPM.GetFormattedBytes(extractedSize)} of {FPM.GetFormattedBytes(totalSize)}";
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

        private void RemoveTask()
        {
            string infoFile = Path.Combine(
                FPM.SourcePath, "Components", workingComponent["path"], $"{workingComponent["title"]}.txt"
            );
            string[] infoText = File.ReadAllLines(infoFile);

            long removedFiles = 0;
            long totalFiles = infoText.Length - 1;
            long totalSize = long.Parse(workingComponent["size"]);

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
                        $"[{(int)((double)totalProgress * 100)}%] Removing component \"{workingComponent["title"]}\"... " +
                        $"{removedFiles} of {totalFiles} files";
                });

                FPM.DeleteFileAndDirectories(filePath);

                removedFiles++;
            }

            FPM.DeleteFileAndDirectories(infoFile);
        }

        private async void FinishDownload()
        {
            if (FPM.DownloadMode == 0)
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
                        IWshShortcut shortcut = new WshShell().CreateShortcut(Path.Combine(path, "Flashpoint.lnk"));
                        shortcut.TargetPath = Path.Combine(FPM.DestinationPath, @"Launcher\Flashpoint.exe");
                        shortcut.WorkingDirectory = Path.Combine(FPM.DestinationPath, @"Launcher");
                        shortcut.Description = "Shortcut to Flashpoint";
                        shortcut.Save();
                    }
                });

                Hide();
                FPM.Main.Hide();

                var finishWindow = new FinishDownload();
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
            
            Close();
        }
    }
}