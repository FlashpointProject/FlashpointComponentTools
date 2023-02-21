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

            ComponentList2.Enabled = true;

            UpdateButton.Enabled = true;
            ChangeButton.Enabled = true;

            ManagerMessage2.Text = "Click on a component to learn more about it.";

            FPM.SyncManager(true);

            ComponentList2.BeforeCheck += ComponentList_BeforeCheck;
            ComponentList2.AfterCheck  += ComponentList2_AfterCheck;
        }

        private void ComponentList2_AfterCheck(object sender, TreeViewEventArgs e)
        {
            FPM.SizeTracker.ToChange = FPM.GetTotalSize(ComponentList2);
        }

        private void ComponentList2_BeforeSelect(object sender, TreeViewCancelEventArgs e)
        {
            if (e.Node.Tag.GetType().ToString().EndsWith("Component"))
            {
                var component = e.Node.Tag as Component;

                DescriptionBox2.Text = "Component Description";
                Description2.Text = component.Description + $" ({FPM.GetFormattedBytes(component.Size)})";
            }
            else
            {
                long categorySize = 0;
                FPM.Iterate(e.Node.Nodes, node =>
                {
                    if (node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        categorySize += (node.Tag as Component).Size;
                    }
                });

                DescriptionBox2.Text = "Category Description";
                Description2.Text = (e.Node.Tag as Category).Description + $" ({FPM.GetFormattedBytes(categorySize)})";
            }

            if (!DescriptionBox2.Visible) DescriptionBox2.Visible = true;
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
                if (!FPM.CheckDependencies(ComponentList2)) return;

                FPM.OperateMode = 1;

                var operationWindow = new Operation();
                operationWindow.ShowDialog();

                FPM.SyncManager();
            }
        }
    }
}