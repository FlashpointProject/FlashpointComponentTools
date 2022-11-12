using System;
using System.Diagnostics;
using System.IO;
using System.Windows.Forms;
using System.Xml;

using Downloader;
using FlashpointInstaller.Common;
using Microsoft.WindowsAPICodePack.Dialogs;

namespace FlashpointInstaller
{
    public partial class Main : Form
    {
        public Main() => InitializeComponent();

        private async void Main_Load(object sender, EventArgs e)
        {
            FPM.VerifyDestinationPath(Path.GetPathRoot(AppDomain.CurrentDomain.BaseDirectory), true);
            
            Stream listStream = await new DownloadService().DownloadFileTaskAsync(FPM.ListURL);
            listStream.Position = 0;

            FPM.XmlTree = new XmlDocument();
            FPM.XmlTree.Load(listStream);
            
            FPM.RecursiveAddToList(FPM.XmlTree.GetElementsByTagName("list")[0], ComponentList.Nodes, true);
        }

        private void Link_LinkClicked(object sender, LinkLabelLinkClickedEventArgs e)
        {
            Process.Start(new ProcessStartInfo("https://bluemaxima.org/flashpoint") { UseShellExecute = true });
        }

        private void ComponentList_BeforeCheck(object sender, TreeViewCancelEventArgs e)
        {
            bool required = e.Node.Tag.GetType().ToString().EndsWith("Component")
                ? (e.Node.Tag as Component).Required
                : (e.Node.Tag as Category).Required;

            if (required && e.Node.Checked) e.Cancel = true;
        }

        private void ComponentList_AfterCheck(object sender, TreeViewEventArgs e)
        {
            FPM.SizeTracker.ToDownload = FPM.GetTotalSize(ComponentList.Nodes);
        }

        private void ComponentList_BeforeSelect(object _, TreeViewCancelEventArgs e)
        {
            Description.Text = e.Node.Tag.GetType().ToString().EndsWith("Component")
                ? (e.Node.Tag as Component).Description
                : (e.Node.Tag as Category).Description;
        }

        private void DestinationPathBrowse_Click(object _, EventArgs e)
        {
            var pathDialog = new CommonOpenFileDialog() { IsFolderPicker = true };
            
            if (pathDialog.ShowDialog() == CommonFileDialogResult.Ok)
            {
                FPM.VerifyDestinationPath(pathDialog.FileName, true);
            }
        }

        private void DestinationPath_KeyPress(object sender, KeyPressEventArgs e)
        {
            if (e.KeyChar == (char)Keys.Return) DownloadButton_Click(this, e);
        }

        private void DownloadButton_Click(object sender, EventArgs e)
        {
            if (FPM.SizeTracker.ToDownload >= 1000000000000)
            {
                var terabyteWarning = MessageBox.Show(
                    "You are about to download OVER A TERABYTE of data. The Complete Offline Archive " +
                    "component is optional and Flashpoint will work without it. Proceed anyway?",
                    "Warning", MessageBoxButtons.YesNo, MessageBoxIcon.Exclamation
                );

                if (terabyteWarning == DialogResult.No) return;
            }

            if (FPM.VerifyDestinationPath(FPM.DestinationPath, false))
            {
                FPM.OperateMode = 0;

                var downloadWindow = new Operation();
                downloadWindow.ShowDialog();
            }
        }
    }
}
