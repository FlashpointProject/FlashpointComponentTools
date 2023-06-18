using System;
using System.Collections.Generic;
using System.Drawing;
using System.IO;
using System.Linq;
using System.Windows.Forms;
using System.Xml;

namespace FlashpointManager
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

            public string InfoFile { get => System.IO.Path.Combine(FPM.SourcePath, "Components", ID); }
            public bool Downloaded { get => File.Exists(InfoFile); }

            // This is used to get around edge cases where check events trigger despite the checkbox value not changing
            public bool Checked { get; set; } = false;

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
                    FPM.ParseError("Root element does not contain URL attribute.");
                }

                // Hash

                string hash = GetAttribute(node, "hash", true);

                if (hash.Length == 8)
                {
                    Hash = hash;
                }
                else
                {
                    FPM.ParseError($"Hash of component \"{Title}\" is invalid.");
                }

                // Size

                long size = 0;
                string stringSize = GetAttribute(node, "install-size", false);

                if (stringSize != "" && !long.TryParse(stringSize, out size))
                {
                    FPM.ParseError($"Size of component \"{Title}\" is not a number.");
                }

                Size = size;

                // LastUpdated

                long lastUpdated;

                if (long.TryParse(GetAttribute(node, "date-modified", true), out lastUpdated))
                {
                    var dateTime = new DateTime(1970, 1, 1, 0, 0, 0, 0).AddSeconds(lastUpdated).ToLocalTime();

                    LastUpdated = dateTime.ToShortDateString();
                }
                else
                {
                    FPM.ParseError($"Date modified of component \"{Title}\" is invalid.");
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
                    FPM.ParseError($"Required {node.Name} attribute \"{attribute}\" was not found");
                }

                return "";
            }
        }

        public static class FPM
        {
            // Pointer to main form
            public static Main Main { get => (Main)Application.OpenForms["Main"]; }
            // The parsed component list XML
            public static XmlDocument XmlTree { get; } = new XmlDocument();

            // Tracks if the manager has been initialized yet
            public static bool Ready { get; set; } = false;

            // Name of configuration file
            public static string ConfigFile { get; set; } = "fpm.cfg";
            // Internet locations of component list XMLs
            public static class RepoXmlTemplates
            {
                public const string Stable = "https://nexus-dev.unstable.life/repository/stable/components.xml";
                public const string Development = "https://nexus-dev.unstable.life/repository/development/components.xml";
            }
            public static string RepoXml { get; set; } = RepoXmlTemplates.Stable;
            // Path to the local Flashpoint copy
            public static string SourcePath { get; set; } = Path.GetFullPath(Path.Combine(Directory.GetCurrentDirectory(), ".."));

            // Controls which components the operate tab adds and removes
            // 0 = Download/remove selected components
            // 1 = Reinstall outdated components
            // 2 = Reinstall broken components
            public static int OperateMode { get; set; } = 0;
            // Controls whether to apply accommodations for offline use
            public static bool OfflineMode { get; set; } = false;
            // Controls whether the update tab is selected at launch
            public static bool OpenUpdateTab { get; set; } = false;
            // Controls whether the launcher should be opened upon the manager closing
            public static bool OpenLauncherOnClose { get; set; } = false;
            // Controls components that will be automatically downloaded at launch
            public static List<string> AutoDownload { get; set; } = new List<string>();

            // Total size of every downloaded component; managed by SyncManager() function
            public static long DownloadedSize { get; set; } = 0;
            // Projected size difference based on changed values of checkboxes
            public static long ModifiedSize { get; set; } = 0;

            // Object providing easy access to certain groups of components; managed by SyncManager() function
            public static class ComponentTracker
            {
                // Contains all downloaded components
                public static List<Component> Downloaded { get; set; } = new List<Component>();
                // Contains all outdated components
                public static List<Component> Outdated   { get; set; } = new List<Component>();
                // Contains all components with missing files
                public static List<Component> Broken     { get; set; } = new List<Component>();
                // Contains all components that no longer exist in the live component repository
                public static List<Component> Deprecated { get; set; } = new List<Component>();
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
                    listNode.Tag  = component;

                    if (component.Required) listNode.ForeColor = Color.FromArgb(255, 96, 96, 96);
                }
                else if (child.Name == "category")
                {
                    var category = new Category(child);

                    listNode.Text = category.Title;
                    listNode.Name = category.ID;
                    listNode.Tag  = category;

                    if (category.Required) listNode.ForeColor = Color.FromArgb(255, 96, 96, 96);
                }

                parent.Add(listNode);
                return listNode;
            }

            // Refreshes component lists and tracker objects with up-to-date information
            public static void SyncManager()
            {
                ComponentTracker.Downloaded.Clear();
                ComponentTracker.Outdated.Clear();
                ComponentTracker.Deprecated.Clear();
                Main.UpdateList.Items.Clear();

                IterateList(Main.ComponentList.Nodes, node =>
                {
                    if (node.Tag.GetType().ToString().EndsWith("Component"))
                    {
                        var component = node.Tag as Component;

                        if (component.Downloaded)
                        {
                            ComponentTracker.Downloaded.Add(component);
                            if (!node.Checked) node.Checked = true;
                        }
                        else
                        {
                            if (OfflineMode) node.ForeColor = Color.FromArgb(255, 96, 96, 96);
                            if (node.Checked) node.Checked = false;
                        }
                    }
                    else
                    {
                        if (OfflineMode) node.ForeColor = Color.FromArgb(255, 96, 96, 96);
                        if (!Ready) node.Checked = node.Checked;
                    }
                });

                DownloadedSize = ComponentTracker.Downloaded.Sum(c => long.Parse(File.ReadLines(c.InfoFile).First().Split(' ')[1]));
                ModifiedSize = 0;

                var componentList = new List<string>();

                IterateXML(XmlTree.GetElementsByTagName("list")[0].ChildNodes, node =>
                {
                    if (node.Name != "component") return;

                    var component = new Component(node);

                    componentList.Add(component.ID);

                    bool downloaded = ComponentTracker.Downloaded.Exists(c => c.ID == component.ID);
                    bool outdated = false;

                    if (downloaded)
                    {
                        string localHash = File.ReadLines(component.InfoFile).First().Split(' ')[0];
                        outdated = localHash != component.Hash && component.ID != "core-database";
                    }
                    else if (component.Required)
                    {
                        outdated = true;
                    }

                    if (outdated)
                    {
                        ComponentTracker.Outdated.Add(component);

                        foreach (string dependID in component.Depends)
                        {
                            if (!(ComponentTracker.Downloaded.Exists(c => c.ID == dependID)
                               || ComponentTracker.Outdated.Exists(c => c.ID == dependID)))
                            {
                                var query = Main.ComponentList.Nodes.Find(dependID, true);
                                if (query.Length > 0) ComponentTracker.Outdated.Add(query[0].Tag as Component);
                            }
                        }
                    }
                });

                ComponentTracker.Outdated = ComponentTracker.Outdated.Distinct().ToList();

                foreach (string filePath in Directory.EnumerateFiles(Path.Combine(SourcePath, "Components")))
                {
                    if (Path.HasExtension(filePath)) continue;

                    string id = Path.GetFileName(filePath).Split('.')[0];

                    if (componentList.Contains(id)) continue;

                    string[] header = File.ReadLines(filePath).First().Split(' ');

                    long size = 0;

                    if (header.Length >= 2 && long.TryParse(header[1], out size))
                    {
                        var component = new Component(XmlTree.SelectSingleNode("//component"))
                        {
                            Title = id,
                            ID    = id,
                            Size  = size,
                            Hash  = header[0]
                        };

                        ComponentTracker.Deprecated.Add(component);
                    }
                }

                long totalSizeChange = 0;

                foreach (var component in ComponentTracker.Outdated)
                {
                    long oldSize = ComponentTracker.Downloaded.Exists(c => c.ID == component.ID)
                        ? long.Parse(File.ReadLines(component.InfoFile).First().Split(' ')[1])
                        : 0;

                    long sizeChange = component.Size - oldSize;
                    totalSizeChange += sizeChange;

                    string displayedSize = GetFormattedBytes(sizeChange);
                    if (displayedSize[0] != '-') displayedSize = "+" + displayedSize;

                    var item = new ListViewItem();
                    item.Text = component.Title;
                    item.SubItems.Add(component.Description);
                    item.SubItems.Add(component.LastUpdated);
                    item.SubItems.Add(displayedSize);

                    Main.UpdateList.Items.Add(item);
                }

                foreach (var component in ComponentTracker.Deprecated)
                {
                    totalSizeChange -= component.Size;

                    var item = new ListViewItem();
                    item.Text = component.Title;
                    item.SubItems.Add("This component is deprecated and can be deleted.");
                    item.SubItems.Add("");
                    item.SubItems.Add(GetFormattedBytes(-component.Size));

                    Main.UpdateList.Items.Add(item);
                }

                Main.ChangeButton.Text = $"Apply changes";
                Main.ChangeButton.Enabled = false;

                Main.UpdateButton.Text = "Install updates";

                if (ComponentTracker.Outdated.Count > 0 || ComponentTracker.Deprecated.Count > 0)
                {
                    Main.UpdateButton.Text += $" ({GetFormattedBytes(totalSizeChange)})";
                    Main.UpdateButton.Enabled = true;
                }
                else
                {
                    Main.UpdateButton.Enabled = false;
                }

                Main.LocationBox.Text = SourcePath;

                switch (RepoXml)
                {
                    case RepoXmlTemplates.Stable:
                        Main.StableRepo.Checked = true;
                        break;
                    case RepoXmlTemplates.Development:
                        Main.DevRepo.Checked = true;
                        break;
                    default:
                        Main.CustomRepo.Checked = true;
                        break;
                }

                Ready = true;
            }

            // Deletes a file as well as any directories made empty by its deletion
            public static void DeleteFileAndDirectories(string file)
            {
                file = Path.GetFullPath(file);

                if (!Path.GetFullPath(file).StartsWith(Path.GetFullPath(SourcePath))) return;

                try
                {
                    File.Delete(file);
                }
                catch (Exception e) when (!(e is DirectoryNotFoundException))
                {
                    GenericError(
                        "Failed to delete the following file:\n" + file + "\n\n" +
                        "Make sure it is not open or being used by another program."
                    );

                    return;
                }

                string folder = Path.GetDirectoryName(file);

                while (folder != Directory.GetParent(SourcePath).ToString())
                {
                    if (Directory.Exists(folder) && !Directory.EnumerateFiles(folder, "*", SearchOption.AllDirectories).Any())
                    {
                        try { Directory.Delete(folder, true); } catch { }
                    }
                    else break;

                    folder = Directory.GetParent(folder).ToString();
                }
            }

            // Checks if specified Flashpoint source path is valid
            public static bool VerifySourcePath(string sourcePath)
            {
                bool isFlashpoint = false;

                IterateXML(XmlTree.GetElementsByTagName("list")[0].ChildNodes, node =>
                {
                    if (isFlashpoint || node.Name != "component") return;

                    if (File.Exists(Path.Combine(sourcePath, "Components", new Component(node).ID)))
                    {
                        isFlashpoint = true;
                    }
                });

                return isFlashpoint;
            }

            // Writes new values to the configuration file
            public static bool WriteConfig(string path, string source)
            {
                try
                {
                    File.WriteAllLines("fpm.cfg", new[] { path, source });
                }
                catch
                {
                    GenericError(
                        "Could not write to configuration file (fpm.cfg)." +
                        "Make sure it is not being used by another program."
                    );

                    return false;
                }

                return true;
            }

            // Checks if any dependencies were not marked for download by the user, and marks them accordingly
            public static bool CheckDependencies(bool alertDepends = true)
            {
                List<string>   requiredDepends = new List<string>();
                List<string>   persistDepends  = new List<string>();
                List<TreeNode> missingDepends  = new List<TreeNode>();

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
                        var component = node.Tag as Component;

                        AddDependencies(ComponentTracker.Downloaded.Exists(c => c.ID == component.ID)
                            ? File.ReadLines(component.InfoFile).First().Split(' ').Skip(2).ToArray()
                            : component.Depends
                        );
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
                            if (ComponentTracker.Downloaded.Exists(depend => depend.ID == component.ID))
                            {
                                persistDepends.Add(component.Title);
                            }
                            else
                            {
                                missingDepends.Add(node);
                            }
                        }
                    }
                });

                if (persistDepends.Count > 0)
                {
                    if (alertDepends)
                    {
                        GenericError(
                            "The following components cannot be removed because one or more components depend on them:\n\n" +
                            string.Join(", ", persistDepends)
                        );
                    }

                    return false;
                }

                if (missingDepends.Count > 0)
                {
                    long missingSize = missingDepends.Select(n => (n.Tag as Component).Size).Sum();
                    var result = DialogResult.Yes;
                    
                    if (alertDepends)
                    {
                        result = MessageBox.Show(
                            "The following dependencies will also be installed:\n\n" +
                            string.Join(", ", missingDepends.Select(n => (n.Tag as Component).Title)) + "\n\n" +
                            $"This will add {GetFormattedBytes(missingSize)} to your download. Is this OK?",
                            "Notice", MessageBoxButtons.YesNo, MessageBoxIcon.Information
                        );
                    }

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
                    if (Math.Abs(bytes) >= unitSize) return (Math.Floor(bytes / unitSize * 10) / 10).ToString("N1") + units[i];
                }

                return "0 bytes";
            }

            // Function for errors unrelated to the component list
            public static void GenericError(string message) => MessageBox.Show(
                message, "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
            );

            // Function for errors related to the component list
            public static void ParseError(string message)
            {
                MessageBox.Show(
                    $"The following error occurred while parsing the component list:\n\n{message}",
                    "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                );

                Environment.Exit(1);
            }
        }
    }
}