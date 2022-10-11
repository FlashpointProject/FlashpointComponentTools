using System;
using System.IO;
using System.Threading.Tasks;
using System.Windows.Forms;

using Downloader;
using IWshRuntimeLibrary;
using SharpCompress.Archives.Zip;
using SharpCompress.Common;
using SharpCompress.Readers;

namespace FlashpointInstaller
{
    public partial class Install : Form
    {
        Main mainForm = (Main)Application.OpenForms["Main"];

        DownloadService downloader = new DownloadService(new DownloadConfiguration());

        Stream stream;
        ZipArchive archive;
        IReader reader;

        bool doneCancelling = false;

        public Install() => InitializeComponent();

        private async void Install_Load(object sender, EventArgs e)
        {
            downloader.DownloadProgressChanged += OnDownloadProgressChanged;

            stream = await downloader.DownloadFileTaskAsync("https://bluepload.unstable.life/selif/flashpointdummy.zip");
            //stream = System.IO.File.OpenRead(@"E:\flashpointdummy.zip");

            if (!downloader.IsCancelled)
            {
                await Task.Run(ExtractTask);
            }
        }

        private void OnDownloadProgressChanged(object sender, DownloadProgressChangedEventArgs e)
        {
            Progress.Invoke((MethodInvoker)delegate
            {
                Progress.Value = (int)((double)e.ReceivedBytesSize / e.TotalBytesToReceive * (Progress.Maximum / 2));
            });

            Info.Invoke((MethodInvoker)delegate
            {
                Info.Text = $"1/2: Downloading archive - {e.ReceivedBytesSize / 1000000}MB of {e.TotalBytesToReceive / 1000000}MB";
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

                    while (!reader.Cancelled && reader.MoveToNextEntry())
                    {
                        if (reader.Entry.IsDirectory)
                        {
                            continue;
                        }

                        reader.WriteEntryToDirectory(mainForm.FolderTextBox.Text, new ExtractionOptions { ExtractFullPath = true, Overwrite = true });

                        extractedSize += reader.Entry.Size;

                        try
                        {
                            Progress.Invoke((MethodInvoker)delegate
                            {
                                Progress.Value = (Progress.Maximum / 2) + (int)((double)extractedSize / totalSize * (Progress.Maximum / 2));
                            });

                            Info.Invoke((MethodInvoker)delegate
                            {
                                Info.Text = $"2/2: Extracting files - {extractedSize / 1000000}MB of {totalSize / 1000000}MB";
                            });
                        }
                        catch { }
                    }

                    if (reader.Cancelled)
                    {
                        Directory.Delete(Path.Combine(mainForm.FolderTextBox.Text, "Flashpoint 11 Infinity"), true);

                        doneCancelling = true;
                    }
                    else
                    {
                        FinishInstallation();
                    }
                }
            }
        }

        private void FinishInstallation()
        {
            if (mainForm.Shortcut.Checked)
            {
                IWshShortcut shortcut = new WshShell().CreateShortcut(Path.Combine(
                    Environment.GetFolderPath(Environment.SpecialFolder.Desktop),
                    "Flashpoint Infinity.lnk"
                ));
                shortcut.TargetPath = Path.Combine(mainForm.FolderTextBox.Text, @"Flashpoint 11 Infinity\Launcher\Flashpoint.exe");
                shortcut.WorkingDirectory = Path.Combine(mainForm.FolderTextBox.Text, @"Flashpoint 11 Infinity\Launcher");
                shortcut.Description = "Shortcut to Flashpoint Infinity";
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