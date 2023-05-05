using System;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using System.Windows.Forms;

using FlashpointManager.Common;

namespace FlashpointManager.src.Forms
{
    public partial class CheckFiles : Form
    {
        public CheckFiles() => InitializeComponent();

        private async void CheckFiles_Load(object sender, EventArgs e)
        {
            FPM.ComponentTracker.Broken.Clear();

            await Task.Run(() =>
            {
                FPM.IterateXML(FPM.XmlTree.GetElementsByTagName("list")[0].ChildNodes, node =>
                {
                    if (node.Name != "component") return;

                    var component = new Component(node);

                    if (!component.Downloaded) return;

                    bool added = false;

                    foreach (string file in File.ReadLines(component.InfoFile).Skip(1))
                    {
                        if (!File.Exists(Path.Combine(FPM.SourcePath, file)))
                        {
                            if (!added)
                            {
                                FPM.ComponentTracker.Broken.Add(component);
                                added = true;
                            }

                            var item = new ListViewItem();
                            item.Text = component.Title;
                            item.SubItems.Add(file);

                            FileList.Invoke((MethodInvoker)delegate { FileList.Items.Add(item); });
                        }
                    }
                });
            });

            FileListLoading.Visible = false;
            FileList.Visible = true;
            
            if (FPM.ComponentTracker.Outdated.Count > 0)
            {
                RedownloadButton.Text = "Cannot redownload components with pending updates";
            }
            else if (FPM.OfflineMode)
            {
                RedownloadButton.Text = "Cannot redownload components in offline mode";
            }
            else
            {
                RedownloadButton.Enabled = FileList.Items.Count > 0;
            }
        }

        private void FileMessage_LinkClicked(object sender, LinkLabelLinkClickedEventArgs e)
        {
            Process.Start(
                new ProcessStartInfo("https://bluemaxima.org/flashpoint/datahub/Troubleshooting_Antivirus_Interference")
                {
                    UseShellExecute = true
                }
            );
        }

        private void RedownloadButton_Click(object sender, EventArgs e)
        {
            FPM.OperateMode = 2;

            new Operate().ShowDialog();

            Close();
        }

        private void CancelButton_Click(object sender, EventArgs e) => Close();
    }
}
