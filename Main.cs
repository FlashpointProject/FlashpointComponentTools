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
        readonly string componentListURL = "http://localhost/components.xml";

        Stream listStream;
        XmlDocument listData;

        public long DownloadSize = 0;
        public string ComponentRootURL;

        public Main() => InitializeComponent();
        
        private async void Main_Load(object sender, EventArgs e)
        {
            SetPath(Path.GetPathRoot(AppDomain.CurrentDomain.BaseDirectory), true);
            
            listStream = await new DownloadService(new DownloadConfiguration()).DownloadFileTaskAsync(componentListURL); listStream.Position = 0;
            listData = new XmlDocument(); listData.Load(listStream);

            ComponentRootURL = listData.GetElementsByTagName("list")[0].Attributes["url"].Value;
            
            void Iterate(XmlNode sourceNode, TreeNodeCollection destNode)
            {
                foreach (XmlNode node in sourceNode.ChildNodes)
                {
                    var listNode = FPM.AddNodeToList(node, destNode);

                    Iterate(node, listNode.Nodes);
                }
            }
            Iterate(listData.GetElementsByTagName("list")[0], ComponentQueue.Nodes);
        }

        private bool SetPath(string path, bool updateText)
        {
            string errorPath;

            if (path.StartsWith(Environment.ExpandEnvironmentVariables("%ProgramW6432%"))
             || path.StartsWith(Environment.GetFolderPath(Environment.SpecialFolder.ProgramFilesX86)))
            {
                errorPath = "Program Files";
            }
            else if (path.StartsWith(Path.GetTempPath().Remove(Path.GetTempPath().Length - 1)))
            {
                errorPath = "Temporary Files";
            }
            else if (path.StartsWith(Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), "OneDrive")))
            {
                errorPath = "OneDrive";
            }
            else
            {
                if (updateText)
                {
                    FolderTextBox.Text = Path.Combine(path, "Flashpoint");
                }

                return true;
            }

            MessageBox.Show(
                $"Flashpoint cannot be installed to the {errorPath} directory! Choose a different folder.",
                "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
            );

            return false;
        }

        private void Link_LinkClicked(object sender, LinkLabelLinkClickedEventArgs e)
        {
            Process.Start(new ProcessStartInfo("https://bluemaxima.org/flashpoint") { UseShellExecute = true });
        }

        private void ComponentList_BeforeCheck(object sender, TreeViewCancelEventArgs e)
        {
            var nodeAttributes = e.Node.Tag as Dictionary<string, string>;

            if (nodeAttributes["disabled"] == "true")
            {
                e.Cancel = true;
            }
        }

        private void ComponentList_AfterCheck(object sender, TreeViewEventArgs e)
        {
            long newDownloadSize = 0;

            void Iterate(TreeNodeCollection parent)
            {
                foreach (TreeNode childNode in parent)
                {
                    var attributes = childNode.Tag as Dictionary<string, string>;

                    if (childNode.Checked && attributes["type"] == "component")
                    {
                        newDownloadSize += int.Parse(attributes["size"]);
                    }

                    Iterate(childNode.Nodes);
                }
            }
            Iterate(ComponentQueue.Nodes);

            DownloadSize = newDownloadSize;
            ComponentSize.Text = FPM.GetFormattedBytes(newDownloadSize);
        }

        private void ComponentQueue_BeforeSelect(object sender, TreeViewCancelEventArgs e)
        {
            Description.Text = (e.Node.Tag as Dictionary<string, string>)["description"];
        }

        private void FolderButton_Click(object sender, EventArgs e)
        {
            var PathDialog = new CommonOpenFileDialog() { IsFolderPicker = true };

            if (PathDialog.ShowDialog() == CommonFileDialogResult.Ok)
            {
                SetPath(PathDialog.FileName, true);
            }
        }

        private void Install_Click(object sender, EventArgs e)
        {
            if (DownloadSize >= 1000000000000)
            {
                var terabyteWarning = MessageBox.Show(
                    "You are about to download OVER A TERABYTE of data. The Complete Offline Archive " +
                    "component is optional and Flashpoint will work without it. Proceed anyway?",
                    "Warning", MessageBoxButtons.YesNo, MessageBoxIcon.Exclamation
                );

                if (terabyteWarning == DialogResult.No) return;
            }

            if (SetPath(FolderTextBox.Text, false))
            {
                var InstallWindow = new Install();
                InstallWindow.ShowDialog();
            }
        }
    }
}
