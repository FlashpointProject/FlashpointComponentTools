using System;
using System.Diagnostics;
using System.IO;
using System.Windows.Forms;

using FlashpointInstaller.Common;

namespace FlashpointInstaller
{
    public partial class FinishDownload : Form
    {
        public FinishDownload() => InitializeComponent();

        private void FinishDownload_Exit(object sender, EventArgs e)
        {
            if (RunOnClose.Checked)
            {
                var flashpointProcess = new Process();
                flashpointProcess.StartInfo.UseShellExecute = true;
                flashpointProcess.StartInfo.FileName = "Flashpoint.exe";
                flashpointProcess.StartInfo.WorkingDirectory = Path.Combine(FPM.DestinationPath, @"Launcher");
                flashpointProcess.Start();
            }

            Environment.Exit(0);
        }
    }
}
