using System.Diagnostics;

namespace FlashpointInstaller
{
    public partial class Finish : Form
    {
        public Finish() => InitializeComponent();

        private void Exit(object sender, EventArgs e)
        {
            if (RunOnClose.Checked)
            {
                Process flashpointProcess = new();
                flashpointProcess.StartInfo.UseShellExecute = true;
                flashpointProcess.StartInfo.FileName = "Flashpoint.exe";
                flashpointProcess.StartInfo.WorkingDirectory = Path.Combine(
                    ((Main)Application.OpenForms["Main"]).FolderTextBox.Text,
                    @"Flashpoint 11 Infinity\Launcher"
                );
                flashpointProcess.Start();
            }

            Environment.Exit(0);
        }
    }
}
