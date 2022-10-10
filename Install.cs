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

        DownloadService downloader = new(new DownloadConfiguration());

        Stream stream;
        ZipArchive archive;
        IReader reader;

        bool finishedWriting = false;

        public Install() => InitializeComponent();

        private async void Install_Load(object sender, EventArgs e)
        {
            downloader.DownloadStarted += OnDownloadStarted;
            downloader.DownloadProgressChanged += OnDownloadProgressChanged;

            stream = await downloader.DownloadFileTaskAsync("http://localhost/Flashpoint%2011%20Infinity.7z");
            //stream = File.OpenRead(@"E:\Flashpoint 11 Infinity.zip");

            if (!downloader.IsCancelled)
            {
                await Task.Run(ExtractTask);
            }
        }

        private void OnDownloadStarted(object? sender, DownloadStartedEventArgs e)
        {
            Info.Invoke((MethodInvoker)delegate
            {
                Info.Text = $"Downloaded 0MB of {e.TotalBytesToReceive / 1000000}MB";
            });
        }

        private void OnDownloadProgressChanged(object? sender, DownloadProgressChangedEventArgs e)
        {
            Progress.Invoke((MethodInvoker)delegate
            {
                Progress.Value = (int)((double)e.ReceivedBytesSize / e.TotalBytesToReceive * (Progress.Maximum / 2));
            });

            Info.Invoke((MethodInvoker)delegate
            {
                Info.Text = $"Downloaded {e.ReceivedBytesSize / 1000000}MB of {e.TotalBytesToReceive / 1000000}MB";
            });
        }

        private void ExtractTask()
        {
            using (archive = ZipArchive.Open(stream))
            {
                using (reader = archive.ExtractAllEntries())
                {
                    int extractedFiles = 0;
                    int totalFiles = archive.Entries.Where(entry => !entry.IsDirectory).Count();
                    
                    while (!reader.Cancelled && reader.MoveToNextEntry())
                    {
                        if (reader.Entry.IsDirectory)
                        {
                            continue;
                        }

                        reader.WriteEntryToDirectory(mainForm.FolderText.Text, new ExtractionOptions { ExtractFullPath = true, Overwrite = true });

                        extractedFiles++;

                        try
                        {
                            Progress.Invoke((MethodInvoker)delegate
                            {
                                Progress.Value = (Progress.Maximum / 2) + ((int)((double)extractedFiles / totalFiles * (Progress.Maximum / 2)));
                            });
                            
                            Info.Invoke((MethodInvoker)delegate
                            {
                                Info.Text = $"Extracted {extractedFiles} of {totalFiles} files";
                            });
                        }
                        catch { }
                    }
                    
                    if (reader.Cancelled)
                    {
                        finishedWriting = true;
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
            if (mainForm.ShortcutsDesktop.Checked)
            {
                string desktopPath = Environment.GetFolderPath(Environment.SpecialFolder.Desktop);

                IWshShortcut shortcut = new WshShell().CreateShortcut(Path.Combine(desktopPath, "Flashpoint Infinity.lnk"));
                shortcut.TargetPath = Path.Combine(mainForm.FolderText.Text, @"Flashpoint 11 Infinity\Launcher\Flashpoint.exe");
                shortcut.Save();
            }

            Finish FinishWindow = new();
            FinishWindow.ShowDialog();
        }

        private async void Cancel_Click(object sender, EventArgs e)
        {
            Cancel.Enabled = false;
            Info.Text = "Cancelling installation...";

            if (downloader.IsBusy)
            {
                downloader.CancelAsync();
            }
            else if (reader != null)
            {
                reader.Cancel();

                await Task.Run(() =>
                {
                    while (!finishedWriting) { }
                });
            }

            Close();
        }
    }
}