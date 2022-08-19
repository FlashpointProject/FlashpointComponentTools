using System.Net.Http.Handlers;

namespace FlashpointInstaller
{
    public partial class Install : Form
    {
        private static ProgressMessageHandler handler = new(new HttpClientHandler() { AllowAutoRedirect = true });
        private HttpClient client = new(handler);

        public Install()
        {
            InitializeComponent();
        }

        private async void Install_Load(object sender, EventArgs e)
        {
            handler.HttpReceiveProgress += Install_ReceiveProgress;

            byte[] fpData = await client.GetByteArrayAsync("http://localhost/Flashpoint%2010.1%20Infinity.7z");
            await File.WriteAllBytesAsync(@"fp.tmp", fpData);

            Close();
        }

        private void Install_ReceiveProgress(object sender, HttpProgressEventArgs e)
        {
            if (this.IsHandleCreated)
            {
                Progress.Invoke((MethodInvoker)delegate
                {
                    Progress.Value = (int)((e.BytesTransferred / e.TotalBytes) * Progress.Maximum);
                });
                
                Info.Invoke((MethodInvoker)delegate
                {
                    Info.Text = $"Downloading... {e.BytesTransferred / 1000000} of {e.TotalBytes / 1000000}";
                });
            }
        }

        private void Cancel_Click(object sender, EventArgs e)
        {
            client.DeleteAsync("http://localhost/Flashpoint%2010.1%20Infinity.7z");
            client.CancelPendingRequests();
            Close();
        }
    }
}