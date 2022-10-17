using System;
using System.Collections.Generic;
using System.IO;
using System.Threading.Tasks;
using System.Windows.Forms;

using Downloader;
using IWshRuntimeLibrary;
using SharpCompress.Archives.Zip;
using SharpCompress.Common;
using SharpCompress.Readers;

using File = System.IO.File;

namespace FlashpointInstaller
{
    public partial class Install : Form
    {
        Main mainForm = (Main)Application.OpenForms["Main"];

        List<Dictionary<string, string>> componentInfo = new List<Dictionary<string, string>>();
        Dictionary<string, string> workingComponent;

        DownloadService downloader = new DownloadService(new DownloadConfiguration());

        Stream stream;
        ZipArchive archive;
        IReader reader;

        long byteProgress = 0;
        long byteTotal = ((Main)Application.OpenForms["Main"]).DownloadSize;

        bool doneCancelling = false;

        public Install() => InitializeComponent();

        private async void Install_Load(object sender, EventArgs e)
        {
            downloader.DownloadProgressChanged += OnDownloadProgressChanged;

            void Iterate(TreeNodeCollection parent)
            {
                foreach (TreeNode childNode in parent)
                {
                    var attributes = childNode.Tag as Dictionary<string, string>;

                    if (attributes["type"] == "component" && childNode.Checked)
                    {
                        componentInfo.Add(childNode.Tag as Dictionary<string, string>);
                    }

                    Iterate(childNode.Nodes);
                }
            }
            Iterate(mainForm.ComponentQueue.Nodes);

            foreach (var component in componentInfo)
            {
                workingComponent = component;

                stream = await downloader.DownloadFileTaskAsync(component["url"]);

                if (!downloader.IsCancelled)
                {
                    Directory.CreateDirectory(mainForm.FolderTextBox.Text);

                    await Task.Run(ExtractTask);
                }
                else { return; }

                byteProgress += int.Parse(component["size"]);
            }

            FinishInstallation();
        }

        private void OnDownloadProgressChanged(object sender, DownloadProgressChangedEventArgs e)
        {
            double currentProgress = (double)e.ReceivedBytesSize / e.TotalBytesToReceive;
            long currentSize = long.Parse(workingComponent["size"]);

            Progress.Invoke((MethodInvoker)delegate
            {
                Progress.Value = (int)((double)
                    ((byteProgress + (currentProgress / 2 * currentSize)) / byteTotal * Progress.Maximum)
                );
            });
            
            Info.Invoke((MethodInvoker)delegate
            {
                Info.Text = 
                    $"Downloading component \"{workingComponent["title"]}\"... " +
                    $"{Math.Truncate(currentProgress * 100)}%";
            });
        }

        private void ExtractTask()
        {
            using (archive = ZipArchive.Open(stream))
            {
                using (reader = archive.ExtractAllEntries())
                {
                    long extractedSize = 0;
                    long totalSize = archive.TotalUncompressSize;

                    string infoPath = Path.Combine(mainForm.FolderTextBox.Text, "Components", workingComponent["path"]);
                    string infoFile = $"{workingComponent["title"]}.txt";

                    Directory.CreateDirectory(infoPath);
                    
                    using (TextWriter writer = File.CreateText(infoPath + infoFile))
                    {
                        writer.WriteLine($"{workingComponent["url"]} {workingComponent["hash"]}");
                    }

                    while (!reader.Cancelled && reader.MoveToNextEntry())
                    {
                        if (reader.Entry.IsDirectory) { continue; }

                        reader.WriteEntryToDirectory(
                            mainForm.FolderTextBox.Text, new ExtractionOptions { ExtractFullPath = true, Overwrite = true }
                        );

                        using (TextWriter writer = File.AppendText(infoPath + infoFile))
                        {
                            writer.WriteLine(reader.Entry.Key);
                        }

                        extractedSize += reader.Entry.Size;

                        try
                        {
                            double currentProgress = (double)extractedSize / totalSize;
                            long currentSize = long.Parse(workingComponent["size"]);

                            Progress.Invoke((MethodInvoker)delegate
                            {
                                Progress.Value = (int)(
                                    (byteProgress + (currentSize / 2) + (currentProgress / 2 * currentSize)) / byteTotal * Progress.Maximum
                                );
                            });

                            Info.Invoke((MethodInvoker)delegate
                            {
                                Info.Text = 
                                    $"Extracting component \"{workingComponent["title"]}\"... " +
                                    $"{Math.Truncate(currentProgress * 100)}%";
                            });
                        }
                        catch { }
                    }

                    if (reader.Cancelled)
                    {
                        Directory.Delete(mainForm.FolderTextBox.Text, true);

                        doneCancelling = true;
                    }
                }
            }
        }

        private void FinishInstallation()
        {
            var shortcutPaths = new List<string>();

            if (mainForm.ShortcutDesktop.Checked)
            {
                shortcutPaths.Add(Environment.GetFolderPath(Environment.SpecialFolder.Desktop));
            }
            if (mainForm.ShortcutStartMenu.Checked)
            {
                shortcutPaths.Add(Environment.GetFolderPath(Environment.SpecialFolder.StartMenu));
            }

            foreach(string path in shortcutPaths)
            {
                IWshShortcut shortcut = new WshShell().CreateShortcut(Path.Combine(path, "Flashpoint.lnk"));
                shortcut.TargetPath = Path.Combine(mainForm.FolderTextBox.Text, @"Launcher\Flashpoint.exe");
                shortcut.WorkingDirectory = Path.Combine(mainForm.FolderTextBox.Text, @"Launcher");
                shortcut.Description = "Shortcut to Flashpoint";
                shortcut.Save();
            }

            Invoke((MethodInvoker)delegate
            {
                Hide();
                mainForm.Hide();

                var FinishWindow = new Finish();
                FinishWindow.ShowDialog();
            });
        }

        private async void Cancel_Click(object sender, EventArgs e)
        {
            Cancel.Enabled = false;

            if (downloader.IsBusy)
            {
                downloader.CancelAsync();
            }
            else if (reader != null)
            {
                reader.Cancel();

                await Task.Run(() => { while (!doneCancelling) { } });
            }

            Close();
        }
    }
}