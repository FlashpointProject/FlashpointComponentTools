using System;
using System.Collections.Generic;
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

        List<Component> addedComponents   = new List<Component>();
        List<Component> removedComponents = new List<Component>();

        DownloadService downloader = new DownloadService();

        Stream stream;
        ZipArchive archive;
        IReader reader;

        long byteProgress = 0;
        long byteTotal = 0;

        public Operate() => InitializeComponent();

        private async void Operation_Load(object sender, EventArgs e)
        {
            TaskbarManager.Instance.SetProgressState(TaskbarProgressBarState.Normal, FPM.Main.Handle);

            downloader.DownloadProgressChanged += OnDownloadProgressChanged;

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
                foreach (var component in FPM.ComponentTracker.ToUpdate)
                {
                    if (FPM.ComponentTracker.Downloaded.Exists(c => c.ID == component.ID))
                    {
                        removedComponents.Add(component);
                    }

                    addedComponents.Add(component);
                }
            }

            byteTotal = removedComponents.Concat(addedComponents).Sum(item => item.Size);

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

                await Task.Run(ExtractComponents);

                byteProgress += component.Size;
            }

            TaskbarManager.Instance.SetProgressState(TaskbarProgressBarState.NoProgress, FPM.Main.Handle);

            Close();
        }

        private void OnDownloadProgressChanged(object sender, DownloadProgressChangedEventArgs e)
        {
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

        private void ExtractComponents()
        {
            using (archive = ZipArchive.Open(stream))
            {
                using (reader = archive.ExtractAllEntries())
                {
                    long extractedSize = 0;
                    long totalSize = archive.TotalUncompressSize;

                    string destPath = Path.Combine(FPM.SourcePath, workingComponent.Path.Replace('/', '\\'));
                    string infoPath = Path.Combine(FPM.SourcePath, "Components");
                    string infoFile = Path.Combine(infoPath, $"{workingComponent.ID}.txt");

                    Directory.CreateDirectory(infoPath);

                    using (TextWriter writer = File.CreateText(infoFile))
                    {
                        string[] header = new[] { workingComponent.Hash, workingComponent.Size.ToString() }.ToArray();

                        writer.WriteLine(string.Join(" ", header));
                    }

                    while (reader.MoveToNextEntry())
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
                double removeProgress = (double)removedFiles / totalFiles;
                double totalProgress = (byteProgress + (removeProgress * totalSize)) / byteTotal;

                FPM.DeleteFileAndDirectories(filePath);
                removedFiles++;

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
            }

            FPM.DeleteFileAndDirectories(infoFile);
        }
    }
}