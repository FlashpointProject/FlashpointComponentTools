using System;
using System.Diagnostics;
using System.IO;
using System.Windows.Forms;
using System.Xml;

using FlashpointInstaller.Common;
using Microsoft.WindowsAPICodePack.Dialogs;

namespace FlashpointInstaller
{
    public partial class Main : Form
    {
        public Main() => InitializeComponent();

        private void Main_Load(object sender, EventArgs e)
        {
            About.Text += $" v{Application.ProductVersion}";

            XmlNodeList rootElements = FPM.XmlTree.GetElementsByTagName("list");

            if (rootElements.Count > 0)
            {
                FPM.RecursiveAddToList(rootElements[0], ComponentList.Nodes, true);
            }
            else
            {
                MessageBox.Show(
                    "An error occurred while parsing the component list XML. Please alert Flashpoint staff ASAP!\n\n" +
                    "Description: Root element was not found",
                    "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                );

                Environment.Exit(1);
            }

            DestinationPath.Text = FPM.DestinationPath;
        }

        private void Link_LinkClicked(object sender, LinkLabelLinkClickedEventArgs e)
        {
            Process.Start(new ProcessStartInfo("https://bluemaxima.org/flashpoint") { UseShellExecute = true });
        }

        public void ComponentList_BeforeCheck(object sender, TreeViewCancelEventArgs e)
        {
            bool required = e.Node.Tag.GetType().ToString().EndsWith("Component")
                ? (e.Node.Tag as Component).Required
                : (e.Node.Tag as Category).Required;

            if (required && e.Node.Checked) e.Cancel = true;
        }

        private void ComponentList_AfterCheck(object sender, TreeViewEventArgs e)
        {
            FPM.SizeTracker.ToDownload = FPM.GetTotalSize(ComponentList);
        }

        private void ComponentList_BeforeSelect(object _, TreeViewCancelEventArgs e)
        {
            if (e.Node.Tag.GetType().ToString().EndsWith("Component"))
            {
                var component = e.Node.Tag as Component;

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
                        categorySize += (node.Tag as Component).Size;
                    }
                });

                DescriptionBox.Text = "Category Description";
                Description.Text = (e.Node.Tag as Category).Description + $" ({FPM.GetFormattedBytes(categorySize)})";
            }

            if (!DescriptionBox.Visible) DescriptionBox.Visible = true;
        }

        private void DestinationPathBrowse_Click(object _, EventArgs e)
        {
            var pathDialog = new CommonOpenFileDialog() { IsFolderPicker = true };
            
            if (pathDialog.ShowDialog() == CommonFileDialogResult.Ok)
            {
                FPM.DestinationPath = Path.Combine(pathDialog.FileName, "Flashpoint");
            }
        }

        private void DestinationPath_KeyPress(object sender, KeyPressEventArgs e)
        {
            if (e.KeyChar == (char)Keys.Return) DownloadButton_Click(this, e);
        }

        private void DownloadButton_Click(object sender, EventArgs e)
        {
            if (!FPM.VerifyDestinationPath(FPM.DestinationPath)) return;

            FPM.CheckDependencies();

            if (!FPM.RedistInstalled)
            {
                var redistDialog = MessageBox.Show(
                    "The Flashpoint launcher requires the Visual C++ 2015 x86 redistributable, which you do not appear to have installed.\n\n" +
                    "It will be installed automatically if you choose to continue.",
                    "Notice", MessageBoxButtons.OKCancel, MessageBoxIcon.Information
                );

                if (redistDialog == DialogResult.Cancel) return;
            }

            var operationWindow = new Operate();
            operationWindow.ShowDialog();
        }
    }
}
