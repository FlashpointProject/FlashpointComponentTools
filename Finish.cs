using System;
using System.Diagnostics;
using System.IO;
using System.Windows.Forms;

namespace FlashpointInstaller
{
    public partial class Finish : Form
    {
        public Finish() => InitializeComponent();

        private void Exit(object sender, EventArgs e)
        {
            if (RunOnClose.Checked)
            {
                var flashpointProcess = new Process();
                flashpointProcess.StartInfo.UseShellExecute = true;
                flashpointProcess.StartInfo.FileName = "Flashpoint.exe";
                flashpointProcess.StartInfo.WorkingDirectory = Path.Combine(
                    ((Main)Application.OpenForms["Main"]).FolderTextBox.Text, @"Launcher"
                );
                flashpointProcess.Start();
            }

            Environment.Exit(0);
        }
    }
}
