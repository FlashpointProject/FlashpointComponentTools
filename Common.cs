using System;
using System.Collections.Generic;
using System.Drawing;
using System.Windows.Forms;
using System.Xml;

namespace FlashpointInstaller
{
    namespace Common
    {
        public static class FPM
        {
            public static Main Main { get { return (Main)Application.OpenForms["Main"]; } }
            public static XmlDocument XmlTree { get; set; }
            public static string ListURL { get; set; }
            public static long DownloadSize { get; set; }
            public static string Path
            {
                get { return Main.FolderTextBox.Text; }
                set { Main.FolderTextBox.Text = value; }
            }

            public static void RecursiveAddToList(XmlNode sourceNode, TreeNodeCollection destNode)
            {
                foreach (XmlNode node in sourceNode.ChildNodes)
                {
                    var listNode = FPM.AddNodeToList(node, destNode);

                    RecursiveAddToList(node, listNode.Nodes);
                }
            }

            public static TreeNode AddNodeToList(XmlNode child, TreeNodeCollection parent)
            {
                var listNode = parent.Add(child.Attributes["title"].Value);
                listNode.Tag = new Dictionary<string, string>
                {
                    { "title", child.Attributes["title"].Value },
                    { "url", GetComponentURL(child) },
                    { "path", GetComponentPath(child) },
                    { "description", child.Attributes["description"].Value },
                    { "type", child.Name },
                    { "disabled", "false" }
                };

                if ((listNode.Tag as Dictionary<string, string>)["type"] == "component")
                {
                    (listNode.Tag as Dictionary<string, string>).Add("size", child.Attributes["size"].Value);
                    (listNode.Tag as Dictionary<string, string>).Add("hash", child.Attributes["hash"].Value);

                    listNode.Checked = bool.Parse(child.Attributes["checked"].Value);
                }

                if ((child.ParentNode.Name == "category" && child.ParentNode.Attributes["required"].Value == "true")
                 || (child.Name == "category" && child.Attributes["required"].Value == "true"))
                {
                    listNode.ForeColor = Color.FromArgb(255, 96, 96, 96);
                    (listNode.Tag as Dictionary<string, string>)["disabled"] = "true";
                }

                return listNode;
            }

            public static long GetEstimatedSize(TreeNodeCollection sourceNodes)
            {
                long size = 0;

                void Iterate(TreeNodeCollection parent)
                {
                    foreach (TreeNode childNode in parent)
                    {
                        var attributes = childNode.Tag as Dictionary<string, string>;

                        if (childNode.Checked && attributes["type"] == "component")
                        {
                            size += int.Parse(attributes["size"]);
                        }

                        Iterate(childNode.Nodes);
                    }
                }
                Iterate(sourceNodes);

                return size;
            }

            public static string GetComponentURL(XmlNode node)
            {
                string path = node.Attributes["url"].Value;

                node = node.ParentNode;

                while (node != null)
                {
                    if (node.Attributes != null)
                    {
                        path = $"{node.Attributes["url"].Value}/{path}";
                    }

                    node = node.ParentNode;
                }

                return path;
            }

            public static string GetComponentPath(XmlNode node)
            {
                string path = "";

                node = node.ParentNode;

                while (node != null && node.Name != "list")
                {
                    if (node.Attributes != null)
                    {
                        path = $"{node.Attributes["url"].Value}\\{path}";
                    }

                    node = node.ParentNode;
                }

                return path;
            }

            public static string GetFormattedBytes(long bytes)
            {
                if (bytes >= 1000000000000)
                {
                    return (Math.Truncate((double)bytes / 100000000000) / 10).ToString("N1") + "TB";
                }
                else if (bytes >= 1000000000)
                {
                    return (Math.Truncate((double)bytes / 100000000) / 10).ToString("N1") + "GB";
                }
                else if (bytes >= 1000000)
                {
                    return (Math.Truncate((double)bytes / 100000) / 10).ToString("N1") + "MB";
                }
                else if (bytes >= 1000)
                {
                    return (Math.Truncate((double)bytes / 100) / 10).ToString("N1") + "KB";
                }
                else
                {
                    return $"{bytes}B";
                }
            }
        }
    }
}