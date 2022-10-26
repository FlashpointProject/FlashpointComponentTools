using System;
using System.Collections.Generic;
using System.IO;
using System.Windows.Forms;

using FlashpointInstaller.Common;
using Microsoft.WindowsAPICodePack.Dialogs;

namespace FlashpointInstaller
{
    public partial class Main : Form
    {
        private void FlashpointPathBrowse_Click(object sender, EventArgs e)
        {
            var PathDialog = new CommonOpenFileDialog() { IsFolderPicker = true };

            if (PathDialog.ShowDialog() != CommonFileDialogResult.Ok) return;
            if (!FPM.SetFlashpointPath(PathDialog.FileName, true)) return;

            ComponentList2.Enabled = true; ManagerMessage2.Visible = false;
            ManagerSizeLabel.Visible = true; ManagerSizeDisplay.Visible = true;

            FPM.Iterate(FPM.Main.ComponentList2.Nodes, node =>
            {
                var attributes = node.Tag as Dictionary<string, string>;

                if (attributes["type"] == "component")
                {
                    string infoPath = Path.Combine(FPM.FlashpointPath, "Components", attributes["path"], $"{attributes["title"]}.txt");

                    if (File.Exists(infoPath))
                    {
                        FPM.InitialComponentInfo.Add(attributes);

                        node.Checked = true;
                    }
                }
            });

            FPM.InitialSize = FPM.GetEstimatedSize(FPM.Main.ComponentList2.Nodes);

            ComponentList2.BeforeCheck += ComponentList_BeforeCheck;
            ComponentList2.AfterCheck  += ComponentList_AfterCheck;
        }

        private void ApplyChanges_Start(object sender, EventArgs e)
        {
            if (FPM.SetFlashpointPath(FPM.FlashpointPath, false))
            {
                FPM.UpdateMode = true;

                var downloadWindow = new Download();
                downloadWindow.ShowDialog();
            }
        }
    }
}