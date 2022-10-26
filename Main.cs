using System;
using System.Collections.Generic;
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
            FPM.ListURL = "http://localhost/components.xml";

            FPM.SetDownloadPath(Path.GetPathRoot(AppDomain.CurrentDomain.BaseDirectory), true);
            
            Stream listStream = await new DownloadService().DownloadFileTaskAsync(FPM.ListURL); listStream.Position = 0;
            FPM.XmlTree = new XmlDocument(); FPM.XmlTree.Load(listStream);
            
            FPM.RecursiveAddToList(FPM.XmlTree.GetElementsByTagName("list")[0], ComponentList.Nodes, true);
            FPM.RecursiveAddToList(FPM.XmlTree.GetElementsByTagName("list")[0], ComponentList2.Nodes, false);
        }

        private void Link_LinkClicked(object sender, LinkLabelLinkClickedEventArgs e)
        {
            Process.Start(new ProcessStartInfo("https://bluemaxima.org/flashpoint") { UseShellExecute = true });
        }

        private void ComponentList_BeforeCheck(object sender, TreeViewCancelEventArgs e)
        {
            var nodeAttributes = e.Node.Tag as Dictionary<string, string>;

            if (bool.Parse(nodeAttributes["disabled"])) e.Cancel = true;
        }

        private void ComponentList_AfterCheck(object sender, TreeViewEventArgs e)
        {
            if (e.Node.TreeView.Name == "ComponentList")
            {
                FPM.DownloadSize = FPM.GetEstimatedSize(ComponentList.Nodes);
            }
            else if (e.Node.TreeView.Name == "ComponentList2")
            {
                FPM.ModifiedSize = FPM.GetEstimatedSize(ComponentList2.Nodes);
            }
        }

        private void ComponentList_BeforeSelect(object sender, TreeViewCancelEventArgs e)
        {
            if (e.Node.TreeView.Name == "ComponentList")
            {
                Description.Text = (e.Node.Tag as Dictionary<string, string>)["description"];
            }
            else if (e.Node.TreeView.Name == "ComponentList2")
            {
                Description2.Text = (e.Node.Tag as Dictionary<string, string>)["description"];
            }
        }

        private void DownloadPathBrowse_Click(object sender, EventArgs e)
        {
            var pathDialog = new CommonOpenFileDialog() { IsFolderPicker = true };
            
            if (pathDialog.ShowDialog() == CommonFileDialogResult.Ok)
            {
                FPM.SetDownloadPath(pathDialog.FileName, true);
            }
        }

        private void DownloadPath_KeyPress(object sender, KeyPressEventArgs e)
        {
            if (e.KeyChar == (char)Keys.Return) Download_Start(this, e);
        }

        private void Download_Start(object sender, EventArgs e)
        {
            if (FPM.DownloadSize >= 1000000000000)
            {
                var terabyteWarning = MessageBox.Show(
                    "You are about to download OVER A TERABYTE of data. The Complete Offline Archive " +
                    "component is optional and Flashpoint will work without it. Proceed anyway?",
                    "Warning", MessageBoxButtons.YesNo, MessageBoxIcon.Exclamation
                );

                if (terabyteWarning == DialogResult.No) return;
            }

            if (FPM.SetDownloadPath(FPM.DownloadPath, false))
            {
                FPM.UpdateMode = false;

                var downloadWindow = new Download();
                downloadWindow.ShowDialog();
            }
        }
    }
}
