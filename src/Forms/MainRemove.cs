using System;
using System.Collections.Generic;
using System.IO;
using System.Threading.Tasks;
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

            FPM.SourcePath2 = pathDialog.FileName;
            if (FPM.SourcePath2 != pathDialog.FileName) return;
        }

        private async void RemoveButton_Click(object sender, EventArgs e)
        {
            if (FPM.VerifySourcePath(FPM.SourcePath2))
            {
                FPM.OperateMode = 3;
                TabControl.Enabled = false;

                await Task.Run(() => {
                    Directory.Delete(FPM.SourcePath2, true);

                    if (FPM.Main.RemoveShortcuts.Checked)
                    {
                        var shortcutPaths = new List<string>()
                        {
                            Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.StartMenu), "Flashpoint.lnk"),
                            Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.Desktop),   "Flashpoint.lnk")
                        };

                        foreach (string path in shortcutPaths) if (File.Exists(path)) File.Delete(path);
                    }
                });

                Hide();

                var finishWindow = new FinishOperation();
                finishWindow.ShowDialog();
            }
        }
    }
}