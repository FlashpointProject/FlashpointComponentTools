using System;
using System.Collections.Generic;
using System.Drawing;
using System.IO;
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
            public static string DownloadPath
            {
                get { return Main.DownloadPath.Text; }
                set { Main.DownloadPath.Text = value; }
            }
            public static string FlashpointPath
            {
                get { return Main.FlashpointPath.Text; }
                set { Main.FlashpointPath.Text = value; }
            }

            private static long downloadSize;
            public static long DownloadSize
            {
                get => downloadSize;
                set
                {
                    Main.DownloadSizeDisplay.Text = GetFormattedBytes(value);
                    downloadSize = value;
                }
            }
            private static long modifiedSize;
            public static long ModifiedSize
            {
                get => modifiedSize;
                set
                {
                    modifiedSize = value;
                    Main.ManagerSizeDisplay.Text = GetFormattedBytes(modifiedSize - InitialSize);
                }
            }
            public static long InitialSize { get; set; }

            public static bool UpdateMode { get; set; } = false;

            public static List<Dictionary<string, string>> InitialComponentInfo { get; set; } = new List<Dictionary<string, string>>();
            public static List<Dictionary<string, string>> AddedComponentInfo   { get; set; } = new List<Dictionary<string, string>>();
            public static List<Dictionary<string, string>> RemovedComponentInfo { get; set; } = new List<Dictionary<string, string>>();

            public static void Iterate(TreeNodeCollection parent, Action<TreeNode> action)
            {
                foreach (TreeNode childNode in parent)
                {
                    action(childNode);

                    Iterate(childNode.Nodes, action);
                }
            }

            public static void RecursiveAddToList(XmlNode sourceNode, TreeNodeCollection destNode, bool setCheckState)
            {
                foreach (XmlNode node in sourceNode.ChildNodes)
                {
                    var listNode = AddNodeToList(node, destNode, setCheckState);

                    RecursiveAddToList(node, listNode.Nodes, setCheckState);
                }
            }

            public static TreeNode AddNodeToList(XmlNode child, TreeNodeCollection parent, bool setCheckState)
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

                    if (setCheckState) listNode.Checked = bool.Parse(child.Attributes["checked"].Value);
                }

                if ((child.ParentNode.Name == "category" && child.ParentNode.Attributes["required"].Value == "true")
                 || (child.Name == "category" && child.Attributes["required"].Value == "true"))
                {
                    listNode.ForeColor = Color.FromArgb(255, 96, 96, 96);
                    (listNode.Tag as Dictionary<string, string>)["disabled"] = "true";
                }

                return listNode;
            }

            public static bool SetDownloadPath(string path, bool updateText)
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
                    if (updateText) DownloadPath = Path.Combine(path, "Flashpoint");

                    return true;
                }

                MessageBox.Show(
                    $"Flashpoint cannot be installed to the {errorPath} directory! Choose a different folder.",
                    "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                );

                return false;
            }

            public static bool SetFlashpointPath(string path, bool updateText)
            {
                bool isFlashpoint = false;

                Iterate(Main.ComponentList2.Nodes, node =>
                {
                    var attributes = node.Tag as Dictionary<string, string>;
                    string infoPath = Path.Combine(path, "Components", attributes["path"], $"{attributes["title"]}.txt");

                    if (File.Exists(infoPath))
                    {
                        isFlashpoint = true;

                        return;
                    }
                });

                if (isFlashpoint)
                {
                    if (updateText) FlashpointPath = path;

                    return true;
                }

                MessageBox.Show($"Flashpoint was not found in this directory!", "Error", MessageBoxButtons.OK, MessageBoxIcon.Error);

                return false;
            }

            public static long GetEstimatedSize(TreeNodeCollection sourceNodes)
            {
                long size = 0;

                Iterate(sourceNodes, node =>
                {
                    var attributes = node.Tag as Dictionary<string, string>;

                    if (node.Checked && attributes["type"] == "component")
                    {
                        size += int.Parse(attributes["size"]);
                    }
                });

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
                else
                {
                    return (Math.Truncate((double)bytes / 100000) / 10).ToString("N1") + "MB";
                }
            }
        }
    }
}