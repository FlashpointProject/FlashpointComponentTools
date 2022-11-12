using System;
using System.Windows.Forms;

using FlashpointInstaller.Common;
using Microsoft.WindowsAPICodePack.Dialogs;

namespace FlashpointInstaller
{
    public partial class Main : Form
    {
        private void TabControl_SelectedIndexChanged(object sender, EventArgs e)
        {
            if (((TabControl)sender).SelectedTab.Name == "ManageTab" && ComponentList2.Nodes.Count == 0)
            {
                FPM.RecursiveAddToList(FPM.XmlTree.GetElementsByTagName("list")[0], ComponentList2.Nodes, false);
            }
        }

        private void SourcePathBrowse_Click(object sender, EventArgs e)
        {
            var pathDialog = new CommonOpenFileDialog() { IsFolderPicker = true };

            if (pathDialog.ShowDialog() != CommonFileDialogResult.Ok) return;
            if (!FPM.VerifySourcePath(pathDialog.FileName, 1)) return;

            ComponentList2.Enabled = true; ManagerMessage2.Visible    = false;
            UpdateButton.Enabled   = true; ManagerSizeLabel.Visible   = true;
            ChangeButton.Enabled   = true; ManagerSizeDisplay.Visible = true;

            FPM.SyncManager(true);

            ComponentList2.BeforeCheck += ComponentList_BeforeCheck;
            ComponentList2.AfterCheck  += ComponentList2_AfterCheck;
        }

        private void ComponentList2_AfterCheck(object sender, TreeViewEventArgs e)
        {
            FPM.SizeTracker.Modified = FPM.GetTotalSize(ComponentList2.Nodes);
        }

        private void ComponentList2_BeforeSelect(object sender, TreeViewCancelEventArgs e)
        {
            Description2.Text = e.Node.Tag.GetType().ToString().EndsWith("Component")
                ? (e.Node.Tag as Component).Description
                : (e.Node.Tag as Category).Description;
        }

        private void UpdateButton_Click(object sender, EventArgs e)
        {
            if (FPM.VerifySourcePath(FPM.SourcePath))
            {
                var updateWindow = new UpdateCheck();
                updateWindow.ShowDialog();
            }
        }

        private void ChangeButton_Click(object sender, EventArgs e)
        {
            if (FPM.VerifySourcePath(FPM.SourcePath))
            {
                FPM.OperateMode = 1;

                var operationWindow = new Operation();
                operationWindow.ShowDialog();

                FPM.SyncManager();
            }
        }
    }
}