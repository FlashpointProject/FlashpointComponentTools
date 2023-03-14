using System;
using System.Diagnostics;
using System.IO;
using System.Windows.Forms;

using FlashpointInstaller.Common;

namespace FlashpointInstaller
{
    public partial class Finish : Form
    {
        public Finish() => InitializeComponent();

        private void FinishDownload_Exit(object sender, EventArgs e)
        {
            if (RunOnClose.Checked)
            {
                new Process() { StartInfo = {
                    UseShellExecute = true,
                    FileName = "Flashpoint.exe",
                    WorkingDirectory = Path.Combine(FPM.Main.DestinationPath.Text, "Launcher")
                }}.Start();
            }

            Environment.Exit(0);
        }
    }
}
