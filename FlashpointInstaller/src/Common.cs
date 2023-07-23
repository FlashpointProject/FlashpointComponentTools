using Microsoft.Win32;
using System;
using System.Collections.Generic;
using System.Drawing;
using System.IO;
using System.Linq;
using System.Net;
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

                long size = 0;
                string stringSize = GetAttribute(node, "install-size", false);

                if (stringSize != "" && !long.TryParse(stringSize, out size))
                {
                    MessageBox.Show(
                        "An error occurred while parsing the component list XML. Please alert Flashpoint staff ASAP!\n\n" +
                        $"Description: Size of component \"{Title}\" is not a number",
                        "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                    );

                    Environment.Exit(1);
                }

                Size = size;

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
                Required = ID.Split('-').FirstOrDefault() == "core";
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
            public static string ListURL { get; set; } = "https://nexus-dev.unstable.life/repository/stable/components.xml";
            // The parsed component list XML
            public static XmlDocument XmlTree { get; set; }
            // WebClient instance used to download files
            public static _WebClient Client { get; } = new _WebClient();

            // Check if Visual C++ 2015 x86 redistributable is installed
            public static bool RedistInstalled
            {
                get => Registry.LocalMachine.OpenSubKey("SOFTWARE\\WOW6432Node\\Microsoft\\VisualStudio\\14.0\\VC\\Runtimes\\X86") != null;
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
                listNode.Checked = child.Name == "component" && !listNode.Name.StartsWith("extra");

                return listNode;
            }

            // Checks if specified Flashpoint destination path is valid
            public static bool VerifyDestinationPath(string path)
            {
                bool alreadyExists = false;

                IterateXML(XmlTree.GetElementsByTagName("list")[0].ChildNodes, node =>
                {
                    if (alreadyExists || node.Name != "component") return;

                    if (File.Exists(Path.Combine(path, "Components", new Component(node).ID)))
                    {
                        alreadyExists = true;
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
            public static bool CheckDependencies()
            {
                List<string> requiredDepends  = new List<string>();
                List<TreeNode> missingDepends = new List<TreeNode>();

                void AddDependencies(string[] depends)
                {
                    requiredDepends.AddRange(depends);

                    foreach (string depend in depends)
                    {
                        var query = Main.ComponentList.Nodes.Find(depend, true);

                        if (query.Length > 0) AddDependencies((query[0].Tag as Component).Depends);
                    }
                }

                // First, fill out a list of dependencies
                IterateList(Main.ComponentList.Nodes, node =>
                {
                    if (node.Checked && node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        AddDependencies((node.Tag as Component).Depends);
                    }
                });

                // Then make sure they're all marked accordingly 
                IterateList(Main.ComponentList.Nodes, node =>
                {
                    if (node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        var component = node.Tag as Component;

                        if (requiredDepends.Any(depend => depend == component.ID) && !node.Checked)
                        {
                            missingDepends.Add(node);
                        }
                    }
                });

                if (missingDepends.Count > 0)
                {
                    long missingSize = missingDepends.Select(n => (n.Tag as Component).Size).Sum();

                    var result = MessageBox.Show(
                        "The following dependencies will also be installed:\n\n" +
                        string.Join(", ", missingDepends.Select(n => (n.Tag as Component).Title)) + "\n\n" +
                        $"This adds an additional {GetFormattedBytes(missingSize)} to your download. Is this OK?",
                        "Notice", MessageBoxButtons.YesNo, MessageBoxIcon.Information
                    );

                    if (result == DialogResult.Yes)
                    {
                        missingDepends.ForEach(d => d.Checked = true);
                    }
                    else
                    {
                        return false;
                    }
                }

                return true;
            }

            // Formats bytes as a human-readable string
            public static string GetFormattedBytes(long bytes)
            {
                string[] units = new[] { " bytes", "KB", "MB", "GB" };
                int i = units.Length;

                while (--i >= 0)
                {
                    double unitSize = Math.Pow(1024, i);
                    if (bytes >= unitSize) return (Math.Round(bytes / unitSize * 10) / 10).ToString("N1") + units[i];
                }

                return "0 bytes";
            }
        }
    }

    // Modified WebClient class with shortened timeout
    public class _WebClient : WebClient
    {
        protected override WebRequest GetWebRequest(Uri address)
        {
            var request = base.GetWebRequest(address);
            request.Timeout = 3000;
            return request;
        }
    }
}