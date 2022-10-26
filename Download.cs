using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
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
    public partial class Download : Form
    {
        List<Dictionary<string, string>> componentInfo = new List<Dictionary<string, string>>();
        Dictionary<string, string> workingComponent;

        DownloadService downloader = new DownloadService(new DownloadConfiguration { OnTheFlyDownload = false });

        Stream stream;
        ZipArchive archive;
        IReader reader;

        long byteProgress = 0;
        long byteTotal = FPM.DownloadSize;

        int cancelStatus = 0;

        public Download() => InitializeComponent();

        private async void Download_Load(object sender, EventArgs e)
        {
            downloader.DownloadProgressChanged += OnDownloadProgressChanged;
            downloader.DownloadFileCompleted += OnDownloadFileCompleted;

            FPM.Iterate(FPM.Main.ComponentList.Nodes, node =>
            {
                var attributes = node.Tag as Dictionary<string, string>;

                if (attributes["type"] == "component" && node.Checked)
                {
                    componentInfo.Add(attributes);
                }
            });

            foreach (var component in componentInfo)
            {
                workingComponent = component;
                stream = await downloader.DownloadFileTaskAsync(component["url"]);
                
                if (cancelStatus != 0) return;

                Directory.CreateDirectory(FPM.DownloadPath);
                await Task.Run(ExtractTask);

                byteProgress += int.Parse(component["size"]);
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

            Progress.Invoke((MethodInvoker)delegate
            {
                Progress.Value = (int)((double)totalProgress * Progress.Maximum);
            });

            Info.Invoke((MethodInvoker)delegate
            {
                Info.Text =
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

                    string infoPath = Path.Combine(FPM.DownloadPath, "Components", workingComponent["path"]);
                    string infoFile = $"{workingComponent["title"]}.txt";

                    Directory.CreateDirectory(infoPath);
                    
                    using (TextWriter writer = File.CreateText(infoPath + infoFile))
                    {
                        writer.WriteLine($"{workingComponent["hash"]} {workingComponent["size"]} {workingComponent["url"]}");
                    }

                    while (cancelStatus == 0 && reader.MoveToNextEntry())
                    {
                        if (reader.Entry.IsDirectory) continue;

                        reader.WriteEntryToDirectory(FPM.DownloadPath, new ExtractionOptions {
                            ExtractFullPath = true, Overwrite = true, PreserveFileTime = true
                        });

                        using (TextWriter writer = File.AppendText(infoPath + infoFile))
                        {
                            writer.WriteLine($"{reader.Entry.Crc} {reader.Entry.Key}");
                        }

                        extractedSize += reader.Entry.Size;

                        double currentProgress = (double)extractedSize / totalSize;
                        long currentSize = long.Parse(workingComponent["size"]);
                        double totalProgress = (byteProgress + (currentSize / 2) + (currentProgress / 2 * currentSize)) / byteTotal;

                        Progress.Invoke((MethodInvoker)delegate
                        {
                            Progress.Value = (int)((double)totalProgress * Progress.Maximum);
                        });

                        Info.Invoke((MethodInvoker)delegate
                        {
                            Info.Text = 
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

        private async void FinishDownload()
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
                    shortcut.TargetPath = Path.Combine(FPM.DownloadPath, @"Launcher\Flashpoint.exe");
                    shortcut.WorkingDirectory = Path.Combine(FPM.DownloadPath, @"Launcher");
                    shortcut.Description = "Shortcut to Flashpoint";
                    shortcut.Save();
                }
            });

            Hide();
            FPM.Main.Hide();

            var finishWindow = new Finish();
            finishWindow.ShowDialog();
        }

        private async void Cancel_Click(object sender, EventArgs e)
        {
            Cancel.Enabled = false;
            cancelStatus = 1;

            await Task.Run(() =>
            { 
                while (cancelStatus != 2) { }

                if (Directory.Exists(FPM.DownloadPath)) Directory.Delete(FPM.DownloadPath, true);
            });
            
            Close();
        }
    }
}