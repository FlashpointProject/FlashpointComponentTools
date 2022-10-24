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

        private void Exit(object sender, EventArgs e)
        {
            if (RunOnClose.Checked)
            {
                var flashpointProcess = new Process();
                flashpointProcess.StartInfo.UseShellExecute = true;
                flashpointProcess.StartInfo.FileName = "Flashpoint.exe";
                flashpointProcess.StartInfo.WorkingDirectory = Path.Combine(FPM.Path, @"Launcher");
                flashpointProcess.Start();
            }

            Environment.Exit(0);
        }
    }
}
