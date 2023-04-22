using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net;
using System.Threading.Tasks;
using System.Windows.Forms;

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

        List<Component> addedComponents   = new List<Component>();
        List<Component> removedComponents = new List<Component>();

        WebClient client = new WebClient();

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

            client.DownloadProgressChanged += OnDownloadProgressChanged;

            if (!FPM.UpdateMode)
            {
                FPM.IterateList(FPM.Main.ComponentList.Nodes, node =>
                {
                    if (node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        var component = node.Tag as Component;

                        if (!node.Checked && FPM.ComponentTracker.Downloaded.Exists(c => c.ID == component.ID))
                        {
                            removedComponents.Add(component);
                        }
                        if (node.Checked && !FPM.ComponentTracker.Downloaded.Exists(c => c.ID == component.ID))
                        {
                            addedComponents.Add(component);
                        }
                    }
                });
            }
            else
            {
                foreach (var component in FPM.ComponentTracker.Outdated)
                {
                    if (FPM.ComponentTracker.Downloaded.Exists(c => c.ID == component.ID))
                    {
                        removedComponents.Add(component);
                    }

                    addedComponents.Add(component);
                }
            }

            byteTotal = FPM.UpdateMode
                ? addedComponents.Sum(c => c.Size)
                : addedComponents.Concat(removedComponents).Sum(c => c.Size);

            foreach (var component in addedComponents)
            {
                workingComponent = component;

                if (component.Size > 0)
                {
                    try
                    {
                        stream = new MemoryStream(await client.DownloadDataTaskAsync(new Uri(component.URL)));
                    }
                    catch (WebException ex) when (ex.Status == WebExceptionStatus.RequestCanceled)
                    {
                        client.Dispose();
                        cancelStatus = 2;
                        return;
                    }
                }

                await Task.Run(ExtractComponents);

                byteProgress += component.Size;
            }

            if (cancelStatus != 0) return;
            CancelButton.Enabled = false;

            foreach (var component in removedComponents)
            {
                workingComponent = component;

                await Task.Run(RemoveComponents);

                byteProgress += component.Size;
            }

            ProgressLabel.Text = "[100%] Finishing up...";

            foreach (var component in addedComponents)
            {
                workingComponent = component;

                await Task.Run(ApplyComponents);
            }

            await Task.Run(DeleteTempDirectory);

            TaskbarManager.Instance.SetProgressState(TaskbarProgressBarState.NoProgress, FPM.Main.Handle);

            if (FPM.AutoDownload.Count == 0) FPM.SyncManager();

            cancelStatus = 2;
            Close();
        }

        private void OnDownloadProgressChanged(object sender, DownloadProgressChangedEventArgs e)
        {
            if (cancelStatus != 0)
            {
                client.CancelAsync();
                return;
            }

            double currentProgress = (double)e.BytesReceived / e.TotalBytesToReceive;
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
                    $"{FPM.GetFormattedBytes(e.BytesReceived)} of {FPM.GetFormattedBytes(e.TotalBytesToReceive)}";
            });

            FPM.Main.Invoke((MethodInvoker)delegate
            {
                TaskbarManager.Instance.SetProgressValue(
                    (int)((double)totalProgress * ProgressMeasure.Maximum), ProgressMeasure.Maximum, FPM.Main.Handle
                );
            });
        }

        private void ExtractComponents()
        {
            string rootPath = Path.Combine(FPM.SourcePath, "Temp");
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

        private void RemoveComponents()
        {
            string[] infoText = File.ReadLines(workingComponent.InfoFile).Skip(1).ToArray();

            long removedFiles = 0;
            long totalFiles = infoText.Length;
            long totalSize = workingComponent.Size;

            foreach (string file in infoText)
            {
                FPM.DeleteFileAndDirectories(Path.Combine(FPM.SourcePath, file));
                removedFiles++;

                double removeProgress = (double)removedFiles / totalFiles;
                double totalProgress = FPM.UpdateMode ? 1 : (byteProgress + (removeProgress * totalSize)) / byteTotal;

                ProgressMeasure.Invoke((MethodInvoker)delegate
                {
                    ProgressMeasure.Value = (int)((double)totalProgress * ProgressMeasure.Maximum);
                });

                ProgressLabel.Invoke((MethodInvoker)delegate
                {
                    string text = FPM.ComponentTracker.Outdated.Exists(c => c.ID == workingComponent.ID)
                        ? $"Removing old version of component \"{workingComponent.Title}\"..."
                        : $"Removing component \"{workingComponent.Title}\"...";

                    ProgressLabel.Text = $"[{(int)((double)totalProgress * 100)}%] {text} {removedFiles} of {totalFiles} files";
                });

                FPM.Main.Invoke((MethodInvoker)delegate
                {
                    TaskbarManager.Instance.SetProgressValue(
                        (int)((double)totalProgress * ProgressMeasure.Maximum), ProgressMeasure.Maximum, FPM.Main.Handle
                    );
                });
            }

            FPM.DeleteFileAndDirectories(workingComponent.InfoFile);
        }

        private void ApplyComponents()
        {
            string tempPath = Path.Combine(FPM.SourcePath, "Temp");
            string tempInfoFile = Path.Combine(tempPath, "Components", $"{workingComponent.ID}.txt");
            string[] infoText = File.ReadLines(tempInfoFile).Skip(1).ToArray();

            foreach (string file in infoText)
            {
                string tempFile = Path.Combine(FPM.SourcePath, "Temp", file);
                string destFile = Path.Combine(FPM.SourcePath, file);

                Directory.CreateDirectory(Path.GetDirectoryName(destFile));
                try { File.Move(tempFile, destFile); } catch { }
            }

            try { File.Move(tempInfoFile, workingComponent.InfoFile); } catch { }
        }

        public static void DeleteTempDirectory()
        {
            string tempPath = Path.Combine(FPM.SourcePath, "Temp");

            if (Directory.Exists(tempPath))
            {
                foreach (string tempFile in Directory.EnumerateFiles(tempPath))
                {
                    try { File.Delete(tempFile); } catch { }
                }

                try { Directory.Delete(tempPath, true); } catch { }
            }
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

                DeleteTempDirectory();
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