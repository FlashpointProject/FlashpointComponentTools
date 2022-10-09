using Downloader;
using SharpCompress.Archives.SevenZip;
using SharpCompress.Common;
using SharpCompress.Readers;

namespace FlashpointInstaller
{
    public partial class Install : Form
    {
        DownloadService downloader = new(new DownloadConfiguration());

        Stream stream;
        SevenZipArchive archive;
        IReader reader;

        string lastEntry = "";
        long extractedSize = 0;

        public Install()
        {
            InitializeComponent();
        }

        private async void Install_Load(object sender, EventArgs e)
        {
            downloader.DownloadStarted += OnDownloadStarted;
            downloader.DownloadProgressChanged += OnDownloadProgressChanged;

            stream = await downloader.DownloadFileTaskAsync("http://localhost/Flashpoint%2011%20Infinity.7z");
            
            await Task.Run(() =>
            {
                using (archive = SevenZipArchive.Open(stream))
                {
                    using (reader = archive.ExtractAllEntries())
                    {
                        reader.EntryExtractionProgress += OnEntryExtractionProgressChanged;
                        reader.WriteAllToDirectory(((Main)Application.OpenForms["Main"]).FolderText.Text, new ExtractionOptions { ExtractFullPath = true, Overwrite = true });
                    }
                }
            });
        }

        private void OnDownloadStarted(object? sender, DownloadStartedEventArgs e)
        {
            Info.Invoke((MethodInvoker)delegate
            {
                Info.Text = $"Downloaded 0MB of {Math.Round((double)e.TotalBytesToReceive / 1000000)}MB";
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

        private void OnEntryExtractionProgressChanged(object? sender, ReaderExtractionEventArgs<IEntry> e)
        {
            if (e.Item.Key != lastEntry)
            {
                extractedSize += e.Item.Size;
                lastEntry = e.Item.Key;
            }

            long extractedSizeExact = extractedSize - (e.Item.Size - e.ReaderProgress.BytesTransferred);

            Progress.Invoke((MethodInvoker)delegate
            {
                Progress.Value = (Progress.Maximum / 2) + ((int)((double)extractedSizeExact / archive.TotalUncompressSize * (Progress.Maximum / 2)));
            });
            
            Info.Invoke((MethodInvoker)delegate
            {
                Info.Text = $"Extracted {extractedSize / 1000000}MB of {archive.TotalUncompressSize / 1000000}MB";
            });
        }

        private void Cancel_Click(object sender, EventArgs e)
        {
            if (downloader.IsBusy)
            {
                downloader.CancelAsync();
            }
            else if (reader != null)
            {
                reader.Cancel();
                archive.Dispose();
                stream.Dispose();
            }

            Close();
        }
    }
}