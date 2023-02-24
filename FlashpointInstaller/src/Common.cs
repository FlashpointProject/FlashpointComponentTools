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
            // Internet location of component list XML
            public static string ListURL { get; set; } = "https://nexus-dev.unstable.life/repository/development/components.xml";
            // The parsed component list XML
            public static XmlDocument XmlTree { get; set; }

            // Check if Visual C++ 2015 x86 redistributable is installed
            public static bool RedistInstalled
            {
                get => Registry.LocalMachine.OpenSubKey("SOFTWARE\\WOW6432Node\\Microsoft\\VisualStudio\\14.0\\VC\\Runtimes\\X86") != null;
            }

            // Pointer to destination path textbox
            private static string destinationPath = Path.Combine(Path.GetPathRoot(AppDomain.CurrentDomain.BaseDirectory), "Flashpoint");
            public static string DestinationPath
            {
                get { return destinationPath; }
                set
                {
                    if (VerifyDestinationPath(value))
                    {
                        destinationPath = value;
                        Main.DestinationPath.Text = destinationPath;
                    }
                }
            }

            // Object for tracking numerous file size sums
            public static class SizeTracker
            {
                private static long toDownload;

                // Tracks total size of the pending Flashpoint download
                public static long ToDownload
                {
                    get => toDownload;
                    set
                    {
                        toDownload = value;
                        Main.InstallButton.Text = $"Install Flashpoint ({GetFormattedBytes(toDownload)})";
                    }
                }
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

            // Checks if specified Flashpoint destination path is valid
            public static bool VerifyDestinationPath(string path)
            {
                bool alreadyExists = false;

                IterateXML(XmlTree.GetElementsByTagName("list")[0].ChildNodes, node =>
                {
                    if (node.Name != "component") return;

                    if (File.Exists(Path.Combine(path, "Components", $"{new Component(node).ID}.txt")))
                    {
                        alreadyExists = true;
                        return;
                    }
                });

                if (alreadyExists)
                {
                    MessageBox.Show(
                        "Flashpoint is already installed to this directory! " +
                        "Choose a different folder or uninstall the existing copy first.",
                        "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                    );

                    return false;
                }

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

            // Checks if any dependencies were not marked for download by the user, and marks them accordingly
            public static void CheckDependencies()
            {
                List<string> requiredDepends = new List<string>();
                List<string> missingDepends  = new List<string>();

                // First, fill out a list of dependencies
                IterateList(Main.ComponentList.Nodes, node =>
                {
                    if (node.Checked && node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        requiredDepends.AddRange((node.Tag as Component).Depends);
                    }
                });

                // Then make sure they're all marked for installation 
                IterateList(Main.ComponentList.Nodes, node =>
                {
                    if (node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        Component component = node.Tag as Component;

                        if (requiredDepends.Any(depend => depend == component.ID) && !node.Checked)
                        {
                            node.Checked = true;
                            missingDepends.Add(component.Title);
                        }
                    }
                });

                if (missingDepends.Count > 0)
                {
                    MessageBox.Show(
                        "The following dependencies will also be installed:\n\n" +
                        string.Join(", ", missingDepends) + "\n\nClick the OK button to proceed.",
                        "Notice", MessageBoxButtons.OK, MessageBoxIcon.Information
                    );
                }
            }

            // Gets total size in bytes of all checked components in the specified TreeView
            public static long GetTotalSize(TreeView sourceTree)
            {
                long size = 0;

                IterateList(sourceTree.Nodes, node =>
                {
                    if (node.Checked && node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        size += (node.Tag as Component).Size;
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