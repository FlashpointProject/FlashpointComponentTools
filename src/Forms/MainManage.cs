using System;
using System.Windows.Forms;

using FlashpointInstaller.Common;
using Microsoft.WindowsAPICodePack.Dialogs;

namespace FlashpointInstaller
{
    public partial class Main : Form
    {
        private void SourcePathBrowse_Click(object sender, EventArgs e)
        {
            var pathDialog = new CommonOpenFileDialog() { IsFolderPicker = true };
            if (pathDialog.ShowDialog() != CommonFileDialogResult.Ok) return;

            FPM.SourcePath = pathDialog.FileName;
            if (FPM.SourcePath != pathDialog.FileName) return;
        }

        public void ComponentList2_AfterCheck(object sender, TreeViewEventArgs e)
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
                FPM.IterateList(e.Node.Nodes, node =>
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