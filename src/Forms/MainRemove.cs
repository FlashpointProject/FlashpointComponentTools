using System;
using System.Windows.Forms;

using FlashpointInstaller.Common;
using Microsoft.WindowsAPICodePack.Dialogs;

namespace FlashpointInstaller
{
    public partial class Main : Form
    {
        private void SourcePathBrowse2_Click(object sender, EventArgs e)
        {
            var pathDialog = new CommonOpenFileDialog() { IsFolderPicker = true };

            if (pathDialog.ShowDialog() != CommonFileDialogResult.Ok) return;
            if (!FPM.VerifySourcePath(pathDialog.FileName, 2)) return;

            RemoveButton.Enabled = true;
        }

        private void RemoveButton_Click(object sender, EventArgs e)
        {
            if (FPM.VerifySourcePath(FPM.SourcePath2))
            {
                FPM.OperateMode = 3;

                var operationWindow = new Operation();
                operationWindow.ShowDialog();
            }
        }
    }
}