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
                    if (FPM.LauncherPath != "")
                    {
                        string[] path = FPM.LauncherPath.Split('\\');

                        var flashpointProcess = new Process();
                        flashpointProcess.StartInfo.UseShellExecute = true;
                        flashpointProcess.StartInfo.FileName = path[1];
                        flashpointProcess.StartInfo.WorkingDirectory = Path.Combine(FPM.DestinationPath, path[0]);
                        flashpointProcess.Start();
                    }
                    else
                    {
                        Process.Start("explorer.exe", FPM.DestinationPath);
                    }
                }
            }

            Environment.Exit(0);
        }
    }
}
