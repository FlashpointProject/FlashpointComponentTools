using System;
using System.ComponentModel;
using System.Diagnostics;
using System.Drawing;
using System.IO;
using System.Linq;
using System.Windows.Forms;
using System.Xml;

using FlashpointManager.Common;

namespace FlashpointManager
{
    public partial class Main : Form
    {
        public Main() => InitializeComponent();

        private void Main_Load(object sender, EventArgs e)
        {
            this.TabControl.ItemSize = new Size(this.TabControl.Width / this.TabControl.TabCount, 0) - new Size(1, 0);
            Text += $" v{Application.ProductVersion}";

            if (FPM.OfflineMode)
            {
                TabControl.TabPages.RemoveAt(1);
                OfflineIndicator.Visible = true;
            }

            using (XmlNodeList rootElements = FPM.XmlTree.GetElementsByTagName("list"))
            {
                if (rootElements.Count > 0)
                {
                    FPM.RecursiveAddToList(rootElements[0], ComponentList.Nodes);
                }
                else
                {
                    FPM.ParseError("Root element was not found.");
                }
            }

            FPM.SyncManager();
            ComponentList.BeforeCheck += ComponentList_BeforeCheck;
            UpdateList.ItemChecked += UpdateList_ItemChecked;

            if (FPM.AutoDownload.Count > 0)
            {
                foreach (string id in FPM.AutoDownload)
                {
                    var query = ComponentList.Nodes.Find(id, true);
                    if (query.Length > 0) query[0].Checked = true;
                }

                FPM.CheckDependencies(false);

                new Operate() {
                    StartPosition = FormStartPosition.CenterScreen,
                    TopMost = true
                }.ShowDialog();

                Close();
            }

            if (FPM.OpenUpdateTab) TabControl.SelectTab(1);
        }

        public void ComponentList_BeforeCheck(object sender, TreeViewCancelEventArgs e)
        {
            if (e.Node.ForeColor == Color.FromArgb(255, 96, 96, 96) && e.Node.Checked)
            {
                e.Cancel = true;
            }
        }

        public void ComponentList_AfterCheck(object sender, TreeViewEventArgs e)
        {
            if (!e.Node.Tag.GetType().ToString().EndsWith("Component")) return;

            var component = e.Node.Tag as Common.Component;

            if (component.Checked == e.Node.Checked) return;
            component.Checked = e.Node.Checked;

            if (!FPM.Ready) return;
            
            FPM.ModifiedSize += (e.Node.Checked ? 1 : -1) * (FPM.ComponentTracker.Downloaded.Exists(c => c.ID == component.ID)
                ? long.Parse(File.ReadLines(component.InfoFile).First().Split(' ')[1])
                : component.Size);

            ChangeButton.Text = $"Apply changes ({FPM.GetFormattedBytes(FPM.ModifiedSize, true)})";
            ChangeButton.Enabled = true;
        }

        private void ComponentList_BeforeSelect(object sender, TreeViewCancelEventArgs e)
        {
            if (e.Node.Tag.GetType().ToString().EndsWith("Component"))
            {
                var component = e.Node.Tag as Common.Component;

                DescriptionBox.Text = "Component Description";
                Description.Text = component.Description + $" ({FPM.GetFormattedBytes(component.Size)})";
            }
            else
            {
                long categorySize = 0;
                FPM.IterateList(e.Node.Nodes, node =>
                {
                    if (node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        categorySize += (node.Tag as Common.Component).Size;
                    }
                });

                DescriptionBox.Text = "Category Description";
                Description.Text = (e.Node.Tag as Category).Description + $" ({FPM.GetFormattedBytes(categorySize)})";
            }

            if (!DescriptionBox.Visible) DescriptionBox.Visible = true;
        }

        private void ChangeButton_Click(object sender, EventArgs e)
        {
            if (!FPM.CheckDependencies()) return;

            FPM.OperationMode = OperateMode.Modify;

            new Operate().ShowDialog();
        }

        private void UpdateButton_Click(object sender, EventArgs e)
        {
            FPM.OperationMode = OperateMode.Update;

            if (UpdateList.CheckedItems.Count == 0)
            {
                MessageBox.Show("No items are checked.", "Flashpoint Manager", MessageBoxButtons.OK, MessageBoxIcon.Information);
                return;
            }

            new Operate().ShowDialog();
        }

        private void Main_FormClosing(object sender, FormClosingEventArgs e)
        {
            if (FPM.OpenLauncherOnClose)
            {
                new Process() { StartInfo = {
                    UseShellExecute = true,
                    FileName = "Flashpoint.exe",
                    WorkingDirectory = Path.Combine(FPM.SourcePath, "Launcher")
                }}.Start();
            }
        }

        public void UpdateList_ItemChecked(object sender, ItemCheckedEventArgs e)
        {
            if (!FPM.Ready || !UpdateList.Focused) return;

            var tagData = (dynamic)e.Item.Tag;
            var component = tagData.Component as Common.Component;
            long sizeChange = tagData.SizeChange;

            if (component != null) {
                if (e.Item.Checked)
                {
                    FPM.totalSizeChange += sizeChange;
                }
                else if (!e.Item.Checked)
                {
                    FPM.totalSizeChange -= sizeChange;
                }

                component.Checked2 = e.Item.Checked;

                lblTotalUpdates.Text = $"Total updates: {UpdateList.CheckedItems.Count}";
                lblTotalUpdatesSize.Text = $"Total size: {FPM.GetFormattedBytes(FPM.totalSizeChange, true)}";
            }

            UpdateButton.Enabled = UpdateList.CheckedItems.Count > 0;
        }

        private void chkUncheckAll_CheckedChanged(object sender, EventArgs e)
        {
            UpdateList.Focus();

            foreach (ListViewItem item in UpdateList.Items)
            {
                item.Checked = chkUncheckAll.Checked;
            }

            UpdateButton.Enabled = chkUncheckAll.Checked;

            ChangeButton.Focus();
        }

        private void OfflineIndicator_Click(object sender, EventArgs e)
        {

        }
    }
}
