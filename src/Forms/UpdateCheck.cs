using System;
using System.IO;
using System.Linq;
using System.Windows.Forms;

using FlashpointInstaller.Common;

namespace FlashpointInstaller
{
    public partial class UpdateCheck : Form
    {
        private long totalSizeChange = 0;

        public UpdateCheck() => InitializeComponent();

        private void UpdateCheck_Load(object sender, EventArgs e)
        {
            if (FPM.StartupMode == 2)
            {
                string[] config = new string[] { };

                if (File.Exists(FPM.ConfigFile)) config = File.ReadAllLines(FPM.ConfigFile);

                if (config.Length > 0)
                {
                    FPM.SourcePath = config[0];
                    FPM.OperateMode = 2;
                }
                else
                {
                    MessageBox.Show("fpm.cfg was not found or is invalid!", "Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
                    Environment.Exit(1);
                }
            }

            FPM.SyncManager();
            FPM.ComponentTracker.ToUpdate.Clear();

            void AddToQueue(Component component, long oldSize)
            {
                long sizeChange = component.Size - oldSize;
                totalSizeChange += sizeChange;

                string displayedSize = FPM.GetFormattedBytes(sizeChange);
                if (displayedSize[0] != '-') displayedSize = "+" + displayedSize;

                var item = new ListViewItem();
                item.Text = component.Title;
                item.SubItems.Add(component.Description);
                item.SubItems.Add(component.LastUpdated);
                item.SubItems.Add(displayedSize);
                UpdateList.Items.Add(item);

                FPM.ComponentTracker.ToUpdate.Add(component);
            }

            FPM.IterateXML(FPM.XmlTree.GetElementsByTagName("list")[0].ChildNodes, node =>
            {
                if (node.Name != "component") return;

                var component = new Component(node);

                bool update = false;
                long oldSize = 0;

                if (FPM.ComponentTracker.Downloaded.Any(item => item.ID == component.ID))
                {
                    string infoFile = Path.Combine(FPM.SourcePath, "Components", $"{component.ID}.txt");
                    string[] componentData = File.ReadLines(infoFile).First().Split(' ');

                    update = componentData[0] != component.Hash;
                    oldSize = long.Parse(componentData[1]);
                }
                else if (component.ID.StartsWith("required"))
                {
                    update = true;
                }

                if (update)
                {
                    AddToQueue(component, oldSize);

                    foreach (string dependID in component.Depends)
                    {
                        if (!FPM.ComponentTracker.Downloaded.Any(item => item.ID == dependID))
                        {
                            FPM.IterateXML(FPM.XmlTree.GetElementsByTagName("list")[0].ChildNodes, node2 =>
                            {
                                if (node2.Name != "component") return;

                                var component2 = new Component(node2);
                                if (component2.ID == dependID) AddToQueue(component2, 0);
                            });
                        }
                    }
                }
            });

            if (FPM.ComponentTracker.ToUpdate.Count > 0)
            {
                UpdateButton.Enabled = true;
                UpdateButton.Text += $" ({FPM.GetFormattedBytes(totalSizeChange)})";
            }
        }

        private void CloseButton_Click(object sender, EventArgs e)
        {
            Close();
        }

        private void UpdateButton_Click(object sender, EventArgs e)
        {
            if (FPM.VerifySourcePath(FPM.SourcePath))
            {
                FPM.OperateMode = 2;

                var downloadWindow = new Operation();
                downloadWindow.ShowDialog();

                FPM.SyncManager(true);

                Close();
            }
        }
    }
}
