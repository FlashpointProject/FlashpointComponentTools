using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Drawing;
using System.IO;
using System.Linq;
using System.Windows.Forms;
using System.Xml;

namespace FlashpointInstaller
{
    namespace Common
    {
        // Component object definition
        public class Component : Category
        {
            public string URL { get; set; }
            public string Hash { get; set; }
            public long Size { get; set; }
            public string LastUpdated { get; set; }
            public string Path { get; set; }
            public string[] Depends { get; set; } = new string[] { };

            public Component(XmlNode node) : base(node)
            {
                // URL

                XmlNode rootElement = node.OwnerDocument.GetElementsByTagName("list")[0];

                if (rootElement.Attributes != null && rootElement.Attributes["url"] != null)
                {
                    URL = rootElement.Attributes["url"].Value + ID + ".zip";
                }
                else
                {
                    MessageBox.Show(
                        "An error occurred while parsing the component list XML. Please alert Flashpoint staff ASAP!\n\n" +
                        "Description: Root element does not contain URL attribute",
                        "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                    );

                    Environment.Exit(1);
                }

                // Hash

                string hash = GetAttribute(node, "hash", true);

                if (hash.Length == 8)
                {
                    Hash = hash;
                }
                else
                {
                    MessageBox.Show(
                        "An error occurred while parsing the component list XML. Please alert Flashpoint staff ASAP!\n\n" +
                        $"Description: Hash of component \"{Title}\" is invalid",
                        "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                    );

                    Environment.Exit(1);
                }

                // Size

                long size;

                if (long.TryParse(GetAttribute(node, "install-size", true), out size))
                {
                    Size = size;
                }
                else
                {
                    MessageBox.Show(
                        "An error occurred while parsing the component list XML. Please alert Flashpoint staff ASAP!\n\n" +
                        $"Description: Size of component \"{Title}\" is not a number",
                        "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                    );

                    Environment.Exit(1);
                }

                // LastUpdated

                long lastUpdated;

                if (long.TryParse(GetAttribute(node, "date-modified", true), out lastUpdated))
                {
                    var dateTime = new DateTime(1970, 1, 1, 0, 0, 0, 0).AddSeconds(lastUpdated).ToLocalTime();

                    LastUpdated = dateTime.ToShortDateString();
                }
                else
                {
                    MessageBox.Show(
                        "An error occurred while parsing the component list XML. Please alert Flashpoint staff ASAP!\n\n" +
                        $"Description: Modified date of component \"{Title}\" is invalid",
                        "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                    );

                    Environment.Exit(1);
                }

                // Path

                Path = GetAttribute(node, "path", false);

                // Depends

                string depends = GetAttribute(node, "depends", false);

                if (depends.Length > 0) Depends = depends.Split(' ');
            }
        }

        // Category object definition
        public class Category
        {
            public string Title { get; set; }
            public string Description { get; set; }
            public string ID { get; set; }
            public bool Required { get; set; }

            public Category(XmlNode node)
            {
                // ID

                XmlNode workingNode = node.ParentNode;
                string id = GetAttribute(node, "id", true);

                while (workingNode != null && workingNode.Name != "list")
                {
                    if (workingNode.Attributes != null && workingNode.Name != "list")
                    {
                        id = $"{GetAttribute(workingNode, "id", true)}-{id}";
                    }

                    workingNode = workingNode.ParentNode;
                }

                ID = id;

                // Everything else

                Title = GetAttribute(node, "title", true);
                Description = GetAttribute(node, "description", true);
                Required = ID.StartsWith("core");
            }

            protected static string GetAttribute(XmlNode node, string attribute, bool throwError)
            {
                if (node.Attributes != null && node.Attributes[attribute] != null)
                {
                    return node.Attributes[attribute].Value;
                }
                else if (throwError)
                {
                    MessageBox.Show(
                        "An error occurred while parsing the component list XML. Please alert Flashpoint staff ASAP!\n\n" +
                        $"Description: Required {node.Name} attribute \"{attribute}\" was not found",
                        "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                    );

                    Environment.Exit(1);
                }

                return "";
            }
        }

        public static class FPM
        {
            // Pointer to main form
            public static Main Main { get => (Main)Application.OpenForms["Main"]; }
            // The parsed component list XML
            public static XmlDocument XmlTree { get; set; }

            // Name of configuration file
            public static string ConfigFile { get; set; } = "fpm.cfg";
            // Internet location of component list XML
            public static string ListURL { get; set; } = "https://nexus-dev.unstable.life/repository/development/components.xml";
            // Path to the local Flashpoint copy
            public static string SourcePath { get; set; } = Debugger.IsAttached
                ? Path.Combine(Path.GetPathRoot(AppDomain.CurrentDomain.BaseDirectory), "Flashpoint")
                : Path.GetFullPath(Path.Combine(Directory.GetCurrentDirectory(), ".."));

            // Flag to control how operation window will function
            // 0 is for adding/removing components
            // 1 is for updating components
            public static int OperateMode { get; set; } = 0;

            // Controls whether the update tab is selected at launch
            public static bool OpenUpdateTab { get; set; } = false;
            // Controls which component (if any) will be automatically downloaded at launch
            public static string AutoDownload { get; set; } = "";

            // Object for tracking numerous file size sums
            public static class SizeTracker
            {
                // Tracks total size of components available locally
                public static long Downloaded { get; set; }
                // Tracks size difference from checking/unchecking components in the manager tab
                private static long toChange;
                public static long ToChange
                {
                    get => toChange;
                    set {
                        toChange = value;
                        Main.ChangeButton.Text = $"Apply changes ({GetFormattedBytes(toChange - Downloaded)})";
                    }
                }
            }

            // Object for tracking information about certain groups of components
            public static class ComponentTracker
            {
                // Information about components that are available locally
                public static List<Component> Downloaded { get; set; } = new List<Component>();
                // Information about components that are to be updated or added through the updater
                public static List<Component> ToUpdate   { get; set; } = new List<Component>();
            }

            // Performs an operation on every node in the specified TreeNodeCollection
            public static void IterateList(TreeNodeCollection parent, Action<TreeNode> action)
            {
                foreach (TreeNode childNode in parent)
                {
                    action(childNode);

                    IterateList(childNode.Nodes, action);
                }
            }

            // Performs an operation on every node in the specified XmlNodeList
            public static void IterateXML(XmlNodeList parent, Action<XmlNode> action)
            {
                foreach (XmlNode childNode in parent)
                {
                    action(childNode);

                    IterateXML(childNode.ChildNodes, action);
                }
            }

            // Calls the AddNodeToList method on every child of the specified XML node
            public static void RecursiveAddToList(XmlNode sourceNode, TreeNodeCollection destNode)
            {
                foreach (XmlNode node in sourceNode.ChildNodes)
                {
                    var listNode = AddNodeToList(node, destNode);

                    RecursiveAddToList(node, listNode.Nodes);
                }
            }

            // Formats an XML node as a TreeView node and adds it to the specified TreeView 
            public static TreeNode AddNodeToList(XmlNode child, TreeNodeCollection parent)
            {
                TreeNode listNode = new TreeNode();

                // Add properties to TreeNode based on the XML element
                // (I can use the dynamic type to prevent redundancy, but I noticed it makes the application load significantly slower)
                if (child.Name == "component")
                {
                    var component = new Component(child);

                    listNode.Text = component.Title;
                    listNode.Name = component.ID;

                    if (component.Required)
                    {
                        listNode.ForeColor = Color.FromArgb(255, 96, 96, 96);
                    }

                    listNode.Tag = component;
                }
                else if (child.Name == "category")
                {
                    var category = new Category(child);

                    listNode.Text = category.Title;
                    listNode.Name = category.ID;

                    if (category.Required)
                    {
                        listNode.ForeColor = Color.FromArgb(255, 96, 96, 96);
                    }

                    listNode.Tag = category;
                }

                parent.Add(listNode);

                // Initialize checkbox
                // (the Checked attribute needs to be explicitly set or else the checkbox won't appear)
                listNode.Checked = false;

                return listNode;
            }

            // Refreshes tracker objects with up-to-date information
            public static void SyncManager(bool updateLists = false)
            {
                ComponentTracker.Downloaded.Clear();
                ComponentTracker.ToUpdate.Clear();

                IterateList(Main.ComponentList.Nodes, node =>
                {
                    if (node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        var component = node.Tag as Component;
                        string infoPath = Path.Combine(SourcePath, "Components", $"{component.ID}.txt");

                        if (File.Exists(infoPath))
                        {
                            ComponentTracker.Downloaded.Add(component);
                            if (updateLists) node.Checked = true;
                        }
                    }
                });

                if (updateLists)
                {
                    Main.UpdateList.Items.Clear();
                    Main.UpdateButton.Text = "Install updates";
                    Main.UpdateButton.Enabled = false;
                }

                SizeTracker.Downloaded = GetTotalSize(Main.ComponentList);
                SizeTracker.ToChange = SizeTracker.Downloaded;
            }

            // Deletes a file as well as any directories made empty by its deletion
            public static void DeleteFileAndDirectories(string file)
            {
                try
                {
                    if (File.Exists(file)) File.Delete(file);

                    string folder = Path.GetDirectoryName(file);

                    while (folder != SourcePath)
                    {
                        if (Directory.Exists(folder) && !Directory.EnumerateFileSystemEntries(folder).Any())
                        {
                            Directory.Delete(folder, false);
                        }
                        else break;

                        folder = Directory.GetParent(folder).ToString();
                    }
                }
                catch { }
            }

            // Checks if specified Flashpoint source path is valid
            public static void VerifySourcePath()
            {
                bool isFlashpoint = false;

                IterateXML(XmlTree.GetElementsByTagName("list")[0].ChildNodes, node =>
                {
                    if (isFlashpoint || node.Name != "component") return;

                    var component = new Component(node);
                    string infoPath = Path.Combine(SourcePath, "Components", $"{component.ID}.txt");

                    if (File.Exists(infoPath))
                    {
                        isFlashpoint = true;
                    }
                });

                if (!isFlashpoint)
                {
                    MessageBox.Show(
                        "The Flashpoint directory specified in fpm.cfg is invalid!",
                        "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                    );

                    Environment.Exit(1);
                }
            }

            // Checks if any dependencies were not marked for download by the user, and marks them accordingly
            public static bool CheckDependencies(TreeView sourceTree)
            {
                List<string> requiredDepends = new List<string>();
                List<string> persistDepends  = new List<string>();
                List<string> missingDepends  = new List<string>();

                // First, fill out a list of dependencies
                IterateList(sourceTree.Nodes, node =>
                {
                    if (node.Checked && node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        var component = node.Tag as Component;
                        string infoPath = Path.Combine(SourcePath, "Components", $"{component.ID}.txt");

                        if (File.Exists(infoPath))
                        {
                            requiredDepends.AddRange(File.ReadLines(infoPath).First().Split(' ').Skip(2).ToArray());
                        }
                        else
                        {
                            requiredDepends.AddRange((node.Tag as Component).Depends);
                        }
                    }
                });

                // Then make sure they're all marked for installation 
                IterateList(sourceTree.Nodes, node =>
                {
                    if (node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        var component = node.Tag as Component;

                        if (requiredDepends.Any(depend => depend == component.ID) && !node.Checked)
                        {
                            node.Checked = true;

                            if (ComponentTracker.Downloaded.Any(depend => depend.ID == component.ID))
                            {
                                persistDepends.Add(component.Title);
                            }
                            else
                            {
                                missingDepends.Add(component.Title);
                            }
                        }
                    }
                });

                if (persistDepends.Count > 0)
                {
                    MessageBox.Show(
                        "The following components cannot be removed because one or more components depend on them:\n\n" +
                        string.Join(", ", persistDepends), "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                    );

                    return false;
                }

                if (missingDepends.Count > 0)
                {
                    MessageBox.Show(
                        "The following dependencies will also be installed:\n\n" +
                        string.Join(", ", missingDepends) + "\n\nClick the OK button to proceed.",
                        "Notice", MessageBoxButtons.OK, MessageBoxIcon.Information
                    );
                }

                return true;
            }

            // Gets total size in bytes of all checked components in the specified TreeView
            public static long GetTotalSize(TreeView sourceTree)
            {
                long size = 0;

                IterateList(sourceTree.Nodes, node =>
                {
                    if (node.Checked && node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        var component = node.Tag as Component;
                        string infoPath = Path.Combine(SourcePath, "Components", $"{component.ID}.txt");

                        size += File.Exists(infoPath)
                            ? long.Parse(File.ReadLines(infoPath).First().Split(' ')[1])
                            : component.Size;
                    }
                });

                return size;
            }

            // Formats bytes as a human-readable string
            public static string GetFormattedBytes(long bytes)
            {
                if (Math.Abs(bytes) >= 1000000000)
                {
                    return (Math.Truncate((double)bytes / 100000000) / 10).ToString("N1") + "GB";
                }
                else if (Math.Abs(bytes) >= 1000000)
                {
                    return (Math.Truncate((double)bytes / 100000) / 10).ToString("N1") + "MB";
                }
                else
                {
                    return (Math.Truncate((double)bytes / 100) / 10).ToString("N1") + "KB";
                }
            }
        }
    }
}