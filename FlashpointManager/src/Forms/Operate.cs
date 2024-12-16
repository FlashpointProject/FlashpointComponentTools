using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Net;
using System.Threading;
using System.Threading.Tasks;
using System.Windows.Forms;

using FlashpointManager.Common;
using Microsoft.WindowsAPICodePack.Taskbar;
using SharpCompress.Archives.Zip;
using SharpCompress.Common;
using SharpCompress.Readers;

namespace FlashpointManager
{
    public partial class Operate : Form
    {
        Component workingComponent;

        List<Component> addedComponents   = new List<Component>();
        List<Component> removedComponents = new List<Component>();

        Stream stream;
        ZipArchive archive;
        IReader reader;

        long byteProgress = 0;
        long byteTotal = 0;

        int cancelStatus = 0;

        private CancellationTokenSource cancellationTokenSource;
        private IProgress<double> progressReporter;

        public Operate()
        {
            InitializeComponent();
            cancellationTokenSource = new CancellationTokenSource();
            progressReporter = new Progress<double>(ReportProgress);

            //FPM.Client.Proxy = new WebProxy("http://127.0.0.1:8888");
        }

        private async void Operation_Load(object sender, EventArgs e)
        {
            var token = cancellationTokenSource.Token;

            TaskbarManager.Instance.SetProgressState(TaskbarProgressBarState.Normal, FPM.Main.Handle);

            FPM.Client.DownloadProgressChanged += OnDownloadProgressChanged;

            if (FPM.OperationMode == OperateMode.Modify)
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
            else if (FPM.OperationMode == OperateMode.Update)
            {
                foreach (var component in FPM.ComponentTracker.Outdated)
                {
                    if (!component.Checked2) continue;
                    if (FPM.ComponentTracker.Downloaded.Exists(c => c.ID == component.ID))
                    {
                        removedComponents.Add(component);
                    }

                    addedComponents.Add(component);
                }

                removedComponents.AddRange(FPM.ComponentTracker.Deprecated.Where(c => c.Checked2));
            }
            else
            {
                foreach (var component in FPM.ComponentTracker.Broken)
                {
                    removedComponents.Add(component);
                    addedComponents.Add(component);
                }
            }

            byteTotal = FPM.OperationMode != OperateMode.Modify
                ? addedComponents.Sum(c => c.Size)
                : addedComponents.Concat(removedComponents).Sum(c => c.Size);

            foreach (var component in addedComponents)
            {
                workingComponent = component;

                if (component.Size > 0)
                {
                    while (true)
                    {
                        try
                        {
                            using (var ctr = token.Register(() => FPM.Client.CancelAsync()))
                            {
                                stream = new MemoryStream(await FPM.Client.DownloadDataTaskAsync(component.URL));
                            }
                        }
                        catch(WebException ex) when (ex.Status == WebExceptionStatus.RequestCanceled)
                        {
                            cancelStatus = 2;
                            return;

                        }
                        catch (WebException ex)
                        {
                            var errorResult = MessageBox.Show(
                                $"The {workingComponent.Title} component failed to download.\n\n" +
                                "Click OK to retry, or Cancel to abort the installation.",
                                "Error", MessageBoxButtons.OKCancel, MessageBoxIcon.Error
                            );
                        
                            if (errorResult == DialogResult.OK)
                            {
                                // Retry the download
                                continue;
                            }
                            else
                            {
                                cancelStatus = 2;
                                cancellationTokenSource.Cancel();
                                return;
                            }
                        }
                        break;
                    }
                }

                await Task.Run(() => ExtractComponents(cancellationTokenSource.Token, progressReporter), cancellationTokenSource.Token);

                byteProgress += component.Size;
            }

            if (cancelStatus != 0) return;
            CancelButton.Enabled = false;

            foreach (var component in removedComponents)
            {
                workingComponent = component;

                await Task.Run(() => RemoveComponents(cancellationTokenSource.Token, progressReporter), cancellationTokenSource.Token);

                byteProgress += component.Size;
            }

            ProgressLabel.Text = "[100%] Finishing up...";

            foreach (var component in addedComponents)
            {
                workingComponent = component;

                await Task.Run(() => ApplyComponents(cancellationTokenSource.Token), cancellationTokenSource.Token);
            }

            TaskbarManager.Instance.SetProgressState(TaskbarProgressBarState.NoProgress, FPM.Main.Handle);

            if (FPM.AutoDownload.Count == 0) FPM.SyncManager();

            cancelStatus = 2;
            Close();
        }

        private void OnDownloadProgressChanged(object sender, DownloadProgressChangedEventArgs e)
        {
            if (cancelStatus != 0)
            {
                FPM.Client.CancelAsync();
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

        private void ExtractComponents(CancellationToken cancellationToken, IProgress<double> progress)
        {
            string rootPath = Path.Combine(FPM.SourcePath, "Temp");
            string infoPath = Path.Combine(rootPath, "Components");
            string infoFile = Path.Combine(infoPath, workingComponent.ID);

            Directory.CreateDirectory(infoPath);

            using (TextWriter writer = File.CreateText(infoFile))
            {
                string[] header = new List<string>
                    { workingComponent.Hash, $"{workingComponent.Size}" }.Concat(workingComponent.Depends).ToArray();

                writer.WriteLine(string.Join(" ", header));
            }

            if (workingComponent.Size == 0) return;
            try
            {
                using (archive = ZipArchive.Open(stream))
                {
                    using (reader = archive.ExtractAllEntries())
                    {
                        long extractedSize = 0;
                        long totalSize = archive.TotalUncompressSize;

                        string destPath = Path.Combine(rootPath, workingComponent.Path.Replace('/', '\\'));

                        while (reader.MoveToNextEntry())
                        {
                            cancellationToken.ThrowIfCancellationRequested();

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

                            progress.Report(totalProgress);

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
                    }
                }
            }
            catch (OperationCanceledException)
            {
                cancelStatus = 2;
            }
        }

        private void RemoveComponents(CancellationToken cancellationToken, IProgress<double> progress)
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
                double totalProgress = FPM.OperationMode != OperateMode.Modify ? 1 : (byteProgress + (removeProgress * totalSize)) / byteTotal;

                progress.Report(totalProgress);

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

        private void ApplyComponents(CancellationToken cancellationToken)
        {
            void MoveDelete(string source, string dest)
            {
                try
                {
                    File.Move(source, dest);
                    FPM.DeleteFileAndDirectories(source);
                }
                catch
                {
                    FPM.GenericError(
                        "Failed to move the following file:\n" + source + "\n\n" +
                        "You will have to move it manually from the Temp folder."
                    );
                }
            }

            string tempPath = Path.Combine(FPM.SourcePath, "Temp");
            string tempInfoFile = Path.Combine(tempPath, "Components", workingComponent.ID);
            string[] infoText = File.ReadLines(tempInfoFile).Skip(1).ToArray();

            foreach (string file in infoText)
            {
                string tempFile = Path.Combine(FPM.SourcePath, "Temp", file);
                string destFile = Path.Combine(FPM.SourcePath, file);

                Directory.CreateDirectory(Path.GetDirectoryName(destFile));

                MoveDelete(tempFile, destFile);

                if (cancellationToken.IsCancellationRequested)
                {
                    cancelStatus = 2;
                    break;
                }
            }

            MoveDelete(tempInfoFile, workingComponent.InfoFile);
        }

        private async void CancelButton_Click(object sender, EventArgs e)
        {
            if (cancelStatus < 1) cancelStatus = 1;

            CancelButton.Enabled = false;
            ProgressLabel.Invoke((MethodInvoker)delegate
            {
                ProgressLabel.Text = "Cancelling...";
            });

            cancellationTokenSource.Cancel();

            await Task.Run(() =>
            {
                while (cancelStatus != 2) { }

                string tempPath = Path.Combine(FPM.SourcePath, "Temp");

                if (Directory.Exists(tempPath))
                {
                    foreach (string tempFile in Directory.EnumerateFiles(tempPath))
                    {
                        try { File.Delete(tempFile); } catch { }
                    }

                    try { Directory.Delete(tempPath, true); } catch { }
                }
            });

            TaskbarManager.Instance.SetProgressState(TaskbarProgressBarState.NoProgress, FPM.Main.Handle);
            Close();
        }

        private void Operation_FormClosing(object sender, FormClosingEventArgs e)
        {
            if (cancelStatus != 2) e.Cancel = true;

            FPM.Client.DownloadProgressChanged -= OnDownloadProgressChanged;
        }

        private void ReportProgress(double value)
        {
            ProgressMeasure.Invoke((MethodInvoker)delegate
            {
                ProgressMeasure.Value = (int)(value * ProgressMeasure.Maximum);
            });
        }
    }
}