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
                Process.Start(Path.Combine(
                    ((Main)Application.OpenForms["Main"]).FolderText.Text,
                    @"Flashpoint 11 Infinity\Launcher\Flashpoint.exe"
                ));
            }

            Environment.Exit(0);
        }
    }
}
