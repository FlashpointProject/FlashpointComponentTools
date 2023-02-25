using System;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using System.Windows.Forms;
using System.Xml;

using FlashpointInstaller.Common;

namespace FlashpointInstaller
{
    public partial class Main : Form
    {
        public Main() => InitializeComponent();

        private void Main_Load(object sender, EventArgs e)
        {
            Text += $" v{Application.ProductVersion}";

            using (XmlNodeList rootElements = FPM.XmlTree.GetElementsByTagName("list"))
            {
                if (rootElements.Count > 0)
                {
                    FPM.RecursiveAddToList(rootElements[0], ComponentList.Nodes);
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
            }

            FPM.SyncManager(true);

            {
                long totalSizeChange = 0;

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

            if (FPM.OpenUpdateTab) TabControl.SelectTab(1);
        }

        public void ComponentList_BeforeCheck(object sender, TreeViewCancelEventArgs e)
        {
            bool required = e.Node.Tag.GetType().ToString().EndsWith("Component")
                ? (e.Node.Tag as Component).Required
                : (e.Node.Tag as Category).Required;

            if (required && e.Node.Checked) e.Cancel = true;
        }

        public void ComponentList_AfterCheck(object sender, TreeViewEventArgs e)
        {
            FPM.SizeTracker.ToChange = FPM.GetTotalSize(ComponentList);
        }

        private void ComponentList_BeforeSelect(object sender, TreeViewCancelEventArgs e)
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

        private void ChangeButton_Click(object sender, EventArgs e)
        {
            if (!FPM.CheckDependencies(ComponentList)) return;

            FPM.OperateMode = 0;

            var operationWindow = new Operate();
            operationWindow.ShowDialog();

            FPM.SyncManager();
        }

        private void UpdateButton_Click(object sender, EventArgs e)
        {
            FPM.OperateMode = 1;

            var downloadWindow = new Operate();
            downloadWindow.ShowDialog();

            FPM.SyncManager(true);
        }

        private async void RemoveButton_Click(object sender, EventArgs e)
        {
            TabControl.Enabled = false;

            await Task.Run(() => {
                foreach (string file in Directory.GetFiles(FPM.SourcePath, "*", SearchOption.AllDirectories))
                {
                    try { FPM.DeleteFileAndDirectories(file); } catch { }
                }

                var shortcutPaths = new string[]
                {
                    Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.StartMenu), "Flashpoint.lnk"),
                    Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.Desktop),   "Flashpoint.lnk")
                };

                foreach (string path in shortcutPaths)
                {
                    try { File.Delete(path); } catch { }
                }
            });

            MessageBox.Show(
                "Flashpoint has been uninstalled from your system.",
                "Uninstallation Complete", MessageBoxButtons.OK, MessageBoxIcon.Information
            );

            Close();
        }
    }
}
