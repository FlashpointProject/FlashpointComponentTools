using System;
using System.Diagnostics;
using System.IO;
using System.Windows.Forms;

using FlashpointInstaller.Common;

namespace FlashpointInstaller
{
    public partial class FinishOperation : Form
    {
        public FinishOperation() => InitializeComponent();

        private void FinishOperation_Load(object sender, EventArgs e)
        {
            if (FPM.OperateMode == 3)
            {
                Text = "Removal Complete";
                Message.Text = "Flashpoint has been successfully removed.";
                RunOnClose.Visible = false;
            }
        }

        private void FinishDownload_Exit(object sender, EventArgs e)
        {
            if (FPM.OperateMode != 3)
            {
                if (RunOnClose.Checked)
                {
                    var flashpointProcess = new Process();
                    flashpointProcess.StartInfo.UseShellExecute = true;
                    flashpointProcess.StartInfo.FileName = "Flashpoint.exe";
                    flashpointProcess.StartInfo.WorkingDirectory = Path.Combine(FPM.DestinationPath, @"Launcher");
                    flashpointProcess.Start();
                }
            }

            Environment.Exit(0);
        }
    }
}
