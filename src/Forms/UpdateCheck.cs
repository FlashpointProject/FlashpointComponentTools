using System.IO;
using System.Linq;
using System.Windows.Forms;
using System.Xml;

using FlashpointInstaller.Common;

namespace FlashpointInstaller
{
    public partial class UpdateCheck : Form
    {
        private long totalSizeChange = 0;

        public UpdateCheck()
        {
            InitializeComponent();
        }

        private void UpdateCheck_Load(object sender, System.EventArgs e)
        {
            FPM.SyncManager();
            FPM.ComponentTracker.ToUpdate.Clear();

            void AddToQueue(Component component, long oldSize)
            {
                long sizeChange = component.Size - oldSize;
                totalSizeChange += sizeChange;

                string displayedSize = FPM.GetFormattedBytes(sizeChange);
                if (displayedSize[0] != '-') displayedSize = "+" + displayedSize;

                ListViewItem item = new ListViewItem();
                item.Text = component.Title;
                item.SubItems.Add(component.Description);
                item.SubItems.Add(displayedSize);
                UpdateList.Items.Add(item);

                FPM.ComponentTracker.ToUpdate.Add(component);
            }

            void IterateXML(XmlNode sourceNode)
            {
                foreach (XmlNode node in sourceNode.ChildNodes)
                {
                    if (node.Name == "component")
                    {
                        Component component = new Component(node);

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
                                    FPM.Iterate(FPM.Main.ComponentList.Nodes, treeNode =>
                                    {
                                        if (treeNode.Tag.GetType().ToString().EndsWith("Component"))
                                        {
                                            Component treeComponent = treeNode.Tag as Component;

                                            if (treeComponent.ID == dependID) AddToQueue(treeComponent, 0);
                                        }
                                    });
                                }
                            }

                            UpdateButton.Enabled = true;
                        }
                    }

                    IterateXML(node);
                }
            }
            IterateXML(FPM.XmlTree.GetElementsByTagName("list")[0]);

            UpdateSizeDisplay.Text = FPM.GetFormattedBytes(totalSizeChange);
        }

        private void CloseButton_Click(object sender, System.EventArgs e)
        {
            Close();
        }

        private void UpdateButton_Click(object sender, System.EventArgs e)
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
