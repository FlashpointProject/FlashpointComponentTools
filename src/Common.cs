using Microsoft.Win32;
using System;
using System.Collections.Generic;
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
            // Pointer to update check form
            public static UpdateCheck UpdateCheck { get => (UpdateCheck)Application.OpenForms["UpdateCheck"]; }
            // Internet location of component list XML
            public static string ListURL { get; set; } = "https://nexus-dev.unstable.life/repository/development/components.xml";
            // The parsed component list XML
            public static XmlDocument XmlTree { get; set; }

            // Name of configuration file
            public static string ConfigFile { get => "fpm.cfg"; }

            // Check if Visual C++ 2015 x86 redistributable is installed
            public static bool RedistInstalled
            {
                get => Registry.LocalMachine.OpenSubKey("SOFTWARE\\WOW6432Node\\Microsoft\\VisualStudio\\14.0\\VC\\Runtimes\\X86") != null;
            }

            // Pointer to destination path textbox
            private static string destinationPath = "";
            public static string DestinationPath
            {
                get { return destinationPath; }
                set
                {
                    if (VerifyDestinationPath(value))
                    {
                        destinationPath = value;

                        if (StartupMode != 2)
                        {
                            Main.DestinationPath.Text = destinationPath;
                        }
                    }
                }
            }
            // Pointer to source path textbox in Manage tab
            private static string sourcePath = "";
            public static string SourcePath
            {
                get { return sourcePath; }
                set
                {
                    if (VerifySourcePath(value))
                    {
                        sourcePath = value;

                        if (StartupMode != 2)
                        {
                            Main.SourcePath.Text = sourcePath;

                            if (!Main.ComponentList2.Enabled)
                            {
                                Main.ComponentList2.Enabled = true;

                                Main.UpdateButton.Enabled = true;
                                Main.ChangeButton.Enabled = true;

                                Main.ManagerMessage2.Text = "Click on a component to learn more about it.";

                                Main.ComponentList2.BeforeCheck += Main.ComponentList_BeforeCheck;
                                Main.ComponentList2.AfterCheck += Main.ComponentList2_AfterCheck;
                            }

                            SyncManager(true);
                        }
                    }
                }
            }
            // Pointer to source path textbox in Remove tab
            private static string sourcePath2 = "";
            public static string SourcePath2
            {
                get { return sourcePath2; }
                set
                {
                    if (VerifySourcePath(value))
                    {
                        sourcePath2 = value;

                        if (StartupMode != 2)
                        {
                            Main.SourcePath2.Text = sourcePath2;
                            Main.RemoveButton.Enabled = true;
                        }
                    }
                }
            }

            // Flag to control how the manager will start
            // 0 starts the manager normally
            // 1 starts the manager in the Manage Flashpoint tab
            // 2 starts the manager in the update window and exits once closed
            public static int StartupMode { get; set; } = 0;

            // Flag to control how operation window will function
            // 0 is for downloading Flashpoint
            // 1 is for adding/removing components
            // 2 is for updating components
            public static int OperateMode { get; set; } = 0;

            // Object for tracking numerous file size sums
            public static class SizeTracker
            {
                private static long toDownload;
                private static long toChange;

                // Tracks total size of components available locally
                public static long Downloaded { get; set; }
                // Tracks total size of the pending Flashpoint download
                public static long ToDownload
                {
                    get => toDownload;
                    set
                    {
                        toDownload = value;
                        Main.DownloadButton.Text = $"Download Flashpoint ({GetFormattedBytes(toDownload)})";
                    }
                }
                // Tracks size difference from checking/unchecking components in the manager tab
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
            public static void RecursiveAddToList(XmlNode sourceNode, TreeNodeCollection destNode, bool setCheckState)
            {
                foreach (XmlNode node in sourceNode.ChildNodes)
                {
                    var listNode = AddNodeToList(node, destNode, setCheckState);

                    RecursiveAddToList(node, listNode.Nodes, setCheckState);
                }
            }

            // Formats an XML node as a TreeView node and adds it to the specified TreeView 
            public static TreeNode AddNodeToList(XmlNode child, TreeNodeCollection parent, bool setCheckState)
            {
                TreeNode listNode = new TreeNode();

                // Add properties to TreeNode based on the XML element
                // (I can use the dynamic type to prevent redundancy, but I noticed it makes the application load significantly slower)
                if (child.Name == "component")
                {
                    Component component = new Component(child);

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
                    Category category = new Category(child);

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
                listNode.Checked = setCheckState && child.Name == "component";

                return listNode;
            }

            // Refreshes tracker objects with up-to-date information
            public static void SyncManager(bool setCheckState = false)
            {
                ComponentTracker.Downloaded.Clear();

                IterateXML(XmlTree.GetElementsByTagName("list")[0].ChildNodes, node =>
                {
                    if (node.Name != "component") return;

                    Component component = new Component(node);
                    string infoPath = Path.Combine(SourcePath, "Components", $"{component.ID}.txt");

                    if (File.Exists(infoPath)) ComponentTracker.Downloaded.Add(component);

                    if (setCheckState && StartupMode != 2)
                    {
                        TreeNode[] nodes = Main.ComponentList2.Nodes.Find(component.ID, true);
                        if (nodes.Length > 0) nodes[0].Checked = File.Exists(infoPath);
                    }
                });

                if (StartupMode != 2)
                {
                    SizeTracker.Downloaded = GetTotalSize(Main.ComponentList2);
                    SizeTracker.ToChange = GetTotalSize(Main.ComponentList2);
                }
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

            // Checks if specified Flashpoint destination path is valid
            public static bool VerifyDestinationPath(string path)
            {
                if (!Path.IsPathRooted(path))
                {
                    MessageBox.Show(
                        $"The specified directory is not valid! Choose a different folder.",
                        "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                    );

                    return false;
                }

                string errorPath;
                
                if (path.StartsWith(Environment.ExpandEnvironmentVariables("%ProgramW6432%"))
                 || path.StartsWith(Environment.GetFolderPath(Environment.SpecialFolder.ProgramFilesX86)))
                {
                    errorPath = "Program Files";
                }
                else if (path.StartsWith(Path.GetTempPath().TrimEnd('\\')))
                {
                    errorPath = "Temporary Files";
                }
                else if (path.StartsWith(Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), "OneDrive")))
                {
                    errorPath = "OneDrive";
                }
                else
                {
                    return true;
                }

                MessageBox.Show(
                    $"Flashpoint cannot be installed to the {errorPath} directory! Choose a different folder.",
                    "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                );

                return false;
            }

            // Checks if specified Flashpoint source path is valid
            public static bool VerifySourcePath(string path)
            {
                bool isFlashpoint = false;

                IterateXML(XmlTree.GetElementsByTagName("list")[0].ChildNodes, node =>
                {
                    if (node.Name != "component") return;

                    Component component = new Component(node);
                    string infoPath = Path.Combine(path, "Components", $"{component.ID}.txt");

                    if (File.Exists(infoPath))
                    {
                        isFlashpoint = true;
                        return;
                    }
                });

                if (isFlashpoint)
                {
                    return true;
                }

                MessageBox.Show($"Flashpoint was not found in this directory!", "Error", MessageBoxButtons.OK, MessageBoxIcon.Error);

                return false;
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
                        Component component = node.Tag as Component;
                        string infoPath = Path.Combine(SourcePath, "Components", $"{component.ID}.txt");

                        if (sourceTree.Name == "ComponentList2" && File.Exists(infoPath))
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
                        Component component = node.Tag as Component;

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
                        Component component = node.Tag as Component;
                        string infoPath = Path.Combine(SourcePath, "Components", $"{component.ID}.txt");

                        size += sourceTree.Name == "ComponentList2" && File.Exists(infoPath)
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