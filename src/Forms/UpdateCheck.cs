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

            void IterateXML(XmlNode sourceNode)
            {
                foreach (XmlNode node in sourceNode.ChildNodes)
                {
                    var query = FPM.ComponentTracker.Downloaded.FirstOrDefault(item => item["url"] == FPM.GetComponentURL(node));

                    if (query != null)
                    {
                        string infoFile = Path.Combine(FPM.SourcePath, "Components", query["path"], $"{query["title"]}.txt");
                        string[] componentData = File.ReadLines(infoFile).First().Split(' ');

                        if (componentData[0] != query["hash"])
                        {
                            long sizeChange = long.Parse(query["size"]) - long.Parse(componentData[1]);
                            totalSizeChange += sizeChange;

                            string displayedSize = FPM.GetFormattedBytes(sizeChange);
                            if (displayedSize[0] != '-') displayedSize = "+" + displayedSize;

                            ListViewItem item = new ListViewItem();
                            item.Text = query["title"];
                            item.SubItems.Add(query["description"]);
                            item.SubItems.Add(displayedSize);
                            UpdateList.Items.Add(item);

                            FPM.ComponentTracker.ToUpdate.Add(query);

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
            if (FPM.SetFlashpointPath(FPM.SourcePath, false))
            {
                FPM.DownloadMode = 2;

                var downloadWindow = new Operation();
                downloadWindow.ShowDialog();

                FPM.SyncManager();

                Close();
            }
        }
    }
}
