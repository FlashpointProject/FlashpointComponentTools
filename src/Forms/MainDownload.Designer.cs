namespace FlashpointInstaller
{
    partial class Main
    {
        /// <summary>
        /// Required designer variable.
        /// </summary>
        private System.ComponentModel.IContainer components = null;

        /// <summary>
        /// Clean up any resources being used.
        /// </summary>
        /// <param name="disposing">true if managed resources should be disposed; otherwise, false.</param>
        protected override void Dispose(bool disposing)
        {
            if (disposing && (components != null))
            {
                components.Dispose();
            }
            base.Dispose(disposing);
        }

        #region Windows Form Designer generated code

        /// <summary>
        /// Required method for Designer support - do not modify
        /// the contents of this method with the code editor.
        /// </summary>
        private void InitializeComponent()
        {
            this.components = new System.ComponentModel.Container();
            System.ComponentModel.ComponentResourceManager resources = new System.ComponentModel.ComponentResourceManager(typeof(Main));
            this.Logo = new System.Windows.Forms.PictureBox();
            this.About = new System.Windows.Forms.Label();
            this.DestinationPathBox = new System.Windows.Forms.GroupBox();
            this.DestinationPath = new System.Windows.Forms.TextBox();
            this.DestinationPathBrowse = new System.Windows.Forms.Button();
            this.DownloadButton = new System.Windows.Forms.Button();
            this.Link = new System.Windows.Forms.LinkLabel();
            this.TabControl = new System.Windows.Forms.TabControl();
            this.DownloadTab = new System.Windows.Forms.TabPage();
            this.DownloadMessage2 = new System.Windows.Forms.Label();
            this.DescriptionBox = new System.Windows.Forms.GroupBox();
            this.SizeDisplay = new System.Windows.Forms.Label();
            this.SizeLabel = new System.Windows.Forms.Label();
            this.Description = new System.Windows.Forms.Label();
            this.tableLayoutPanel1 = new System.Windows.Forms.TableLayoutPanel();
            this.ShortcutStartMenu = new System.Windows.Forms.CheckBox();
            this.ShortcutLabel = new System.Windows.Forms.Label();
            this.ShortcutDesktop = new System.Windows.Forms.CheckBox();
            this.TotalSizeDisplay = new System.Windows.Forms.Label();
            this.TotalSizeLabel = new System.Windows.Forms.Label();
            this.DownloadMessage = new System.Windows.Forms.Label();
            this.ComponentList = new RikTheVeggie.TriStateTreeView();
            this.ManageTab = new System.Windows.Forms.TabPage();
            this.ManagerMessage2 = new System.Windows.Forms.Label();
            this.ManagerMessage = new System.Windows.Forms.Label();
            this.TotalSizeDisplay2 = new System.Windows.Forms.Label();
            this.TotalSizeLabel2 = new System.Windows.Forms.Label();
            this.UpdateButton = new System.Windows.Forms.Button();
            this.ChangeButton = new System.Windows.Forms.Button();
            this.SourcePathBox = new System.Windows.Forms.GroupBox();
            this.SourcePath = new System.Windows.Forms.TextBox();
            this.SourcePathBrowse = new System.Windows.Forms.Button();
            this.DescriptionBox2 = new System.Windows.Forms.GroupBox();
            this.SizeDisplay2 = new System.Windows.Forms.Label();
            this.SizeLabel2 = new System.Windows.Forms.Label();
            this.Description2 = new System.Windows.Forms.Label();
            this.ComponentList2 = new RikTheVeggie.TriStateTreeView();
            this.RemoveTab = new System.Windows.Forms.TabPage();
            this.label1 = new System.Windows.Forms.Label();
            this.RemoveShortcuts = new System.Windows.Forms.CheckBox();
            this.SourcePathBox2 = new System.Windows.Forms.GroupBox();
            this.SourcePath2 = new System.Windows.Forms.TextBox();
            this.SourcePathBrowse2 = new System.Windows.Forms.Button();
            this.RemoveButton = new System.Windows.Forms.Button();
            ((System.ComponentModel.ISupportInitialize)(this.Logo)).BeginInit();
            this.DestinationPathBox.SuspendLayout();
            this.TabControl.SuspendLayout();
            this.DownloadTab.SuspendLayout();
            this.DescriptionBox.SuspendLayout();
            this.tableLayoutPanel1.SuspendLayout();
            this.ManageTab.SuspendLayout();
            this.SourcePathBox.SuspendLayout();
            this.DescriptionBox2.SuspendLayout();
            this.RemoveTab.SuspendLayout();
            this.SourcePathBox2.SuspendLayout();
            this.SuspendLayout();
            // 
            // Logo
            // 
            this.Logo.Image = ((System.Drawing.Image)(resources.GetObject("Logo.Image")));
            this.Logo.Location = new System.Drawing.Point(12, 12);
            this.Logo.Name = "Logo";
            this.Logo.Size = new System.Drawing.Size(600, 177);
            this.Logo.SizeMode = System.Windows.Forms.PictureBoxSizeMode.CenterImage;
            this.Logo.TabIndex = 0;
            this.Logo.TabStop = false;
            // 
            // About
            // 
            this.About.AutoSize = true;
            this.About.Location = new System.Drawing.Point(12, 192);
            this.About.Name = "About";
            this.About.Size = new System.Drawing.Size(100, 13);
            this.About.TabIndex = 1;
            this.About.Text = "Flashpoint Manager";
            // 
            // DestinationPathBox
            // 
            this.DestinationPathBox.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.DestinationPathBox.Controls.Add(this.DestinationPath);
            this.DestinationPathBox.Controls.Add(this.DestinationPathBrowse);
            this.DestinationPathBox.Location = new System.Drawing.Point(7, 199);
            this.DestinationPathBox.Name = "DestinationPathBox";
            this.DestinationPathBox.Size = new System.Drawing.Size(576, 49);
            this.DestinationPathBox.TabIndex = 2;
            this.DestinationPathBox.TabStop = false;
            this.DestinationPathBox.Text = "Destination folder:";
            // 
            // DestinationPath
            // 
            this.DestinationPath.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Regular, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.DestinationPath.Location = new System.Drawing.Point(9, 18);
            this.DestinationPath.Name = "DestinationPath";
            this.DestinationPath.Size = new System.Drawing.Size(477, 20);
            this.DestinationPath.TabIndex = 1;
            this.DestinationPath.KeyPress += new System.Windows.Forms.KeyPressEventHandler(this.DestinationPath_KeyPress);
            // 
            // DestinationPathBrowse
            // 
            this.DestinationPathBrowse.Location = new System.Drawing.Point(495, 17);
            this.DestinationPathBrowse.Name = "DestinationPathBrowse";
            this.DestinationPathBrowse.Size = new System.Drawing.Size(73, 22);
            this.DestinationPathBrowse.TabIndex = 0;
            this.DestinationPathBrowse.Text = "Browse";
            this.DestinationPathBrowse.UseVisualStyleBackColor = true;
            this.DestinationPathBrowse.Click += new System.EventHandler(this.DestinationPathBrowse_Click);
            // 
            // DownloadButton
            // 
            this.DownloadButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.DownloadButton.Location = new System.Drawing.Point(300, 256);
            this.DownloadButton.Name = "DownloadButton";
            this.DownloadButton.Size = new System.Drawing.Size(284, 26);
            this.DownloadButton.TabIndex = 2;
            this.DownloadButton.Text = "Download Flashpoint";
            this.DownloadButton.UseVisualStyleBackColor = true;
            this.DownloadButton.Click += new System.EventHandler(this.DownloadButton_Click);
            // 
            // Link
            // 
            this.Link.AutoSize = true;
            this.Link.LinkColor = System.Drawing.Color.FromArgb(((int)(((byte)(4)))), ((int)(((byte)(94)))), ((int)(((byte)(221)))));
            this.Link.Location = new System.Drawing.Point(441, 192);
            this.Link.Name = "Link";
            this.Link.Size = new System.Drawing.Size(171, 13);
            this.Link.TabIndex = 4;
            this.Link.TabStop = true;
            this.Link.Text = "https://bluemaxima.org/flashpoint/";
            this.Link.LinkClicked += new System.Windows.Forms.LinkLabelLinkClickedEventHandler(this.Link_LinkClicked);
            // 
            // TabControl
            // 
            this.TabControl.Controls.Add(this.DownloadTab);
            this.TabControl.Controls.Add(this.ManageTab);
            this.TabControl.Controls.Add(this.RemoveTab);
            this.TabControl.ItemSize = new System.Drawing.Size(150, 20);
            this.TabControl.Location = new System.Drawing.Point(12, 229);
            this.TabControl.Name = "TabControl";
            this.TabControl.SelectedIndex = 0;
            this.TabControl.Size = new System.Drawing.Size(600, 320);
            this.TabControl.SizeMode = System.Windows.Forms.TabSizeMode.Fixed;
            this.TabControl.TabIndex = 5;
            this.TabControl.SelectedIndexChanged += new System.EventHandler(this.TabControl_SelectedIndexChanged);
            // 
            // DownloadTab
            // 
            this.DownloadTab.Controls.Add(this.DownloadMessage2);
            this.DownloadTab.Controls.Add(this.DescriptionBox);
            this.DownloadTab.Controls.Add(this.tableLayoutPanel1);
            this.DownloadTab.Controls.Add(this.TotalSizeDisplay);
            this.DownloadTab.Controls.Add(this.TotalSizeLabel);
            this.DownloadTab.Controls.Add(this.DownloadMessage);
            this.DownloadTab.Controls.Add(this.ComponentList);
            this.DownloadTab.Controls.Add(this.DestinationPathBox);
            this.DownloadTab.Controls.Add(this.DownloadButton);
            this.DownloadTab.Location = new System.Drawing.Point(4, 24);
            this.DownloadTab.Name = "DownloadTab";
            this.DownloadTab.Padding = new System.Windows.Forms.Padding(3);
            this.DownloadTab.Size = new System.Drawing.Size(592, 292);
            this.DownloadTab.TabIndex = 0;
            this.DownloadTab.Text = "Download Flashpoint";
            this.DownloadTab.UseVisualStyleBackColor = true;
            // 
            // DownloadMessage2
            // 
            this.DownloadMessage2.Location = new System.Drawing.Point(410, 44);
            this.DownloadMessage2.Name = "DownloadMessage2";
            this.DownloadMessage2.Size = new System.Drawing.Size(172, 30);
            this.DownloadMessage2.TabIndex = 13;
            this.DownloadMessage2.Text = "Click on a component to learn more about it.";
            // 
            // DescriptionBox
            // 
            this.DescriptionBox.Controls.Add(this.SizeDisplay);
            this.DescriptionBox.Controls.Add(this.SizeLabel);
            this.DescriptionBox.Controls.Add(this.Description);
            this.DescriptionBox.Location = new System.Drawing.Point(411, 100);
            this.DescriptionBox.Name = "DescriptionBox";
            this.DescriptionBox.Padding = new System.Windows.Forms.Padding(10, 8, 10, 10);
            this.DescriptionBox.Size = new System.Drawing.Size(163, 91);
            this.DescriptionBox.TabIndex = 12;
            this.DescriptionBox.TabStop = false;
            this.DescriptionBox.Text = "Component Description";
            // 
            // SizeDisplay
            // 
            this.SizeDisplay.AutoSize = true;
            this.SizeDisplay.Location = new System.Drawing.Point(45, 68);
            this.SizeDisplay.Margin = new System.Windows.Forms.Padding(0);
            this.SizeDisplay.Name = "SizeDisplay";
            this.SizeDisplay.Size = new System.Drawing.Size(38, 13);
            this.SizeDisplay.TabIndex = 9;
            this.SizeDisplay.Text = "0.0MB";
            this.SizeDisplay.Visible = false;
            // 
            // SizeLabel
            // 
            this.SizeLabel.AutoSize = true;
            this.SizeLabel.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Bold, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.SizeLabel.Location = new System.Drawing.Point(10, 68);
            this.SizeLabel.Margin = new System.Windows.Forms.Padding(0);
            this.SizeLabel.Name = "SizeLabel";
            this.SizeLabel.Size = new System.Drawing.Size(35, 13);
            this.SizeLabel.TabIndex = 8;
            this.SizeLabel.Text = "Size:";
            this.SizeLabel.Visible = false;
            // 
            // Description
            // 
            this.Description.Dock = System.Windows.Forms.DockStyle.Top;
            this.Description.Location = new System.Drawing.Point(10, 21);
            this.Description.Name = "Description";
            this.Description.Size = new System.Drawing.Size(143, 44);
            this.Description.TabIndex = 0;
            // 
            // tableLayoutPanel1
            // 
            this.tableLayoutPanel1.ColumnCount = 3;
            this.tableLayoutPanel1.ColumnStyles.Add(new System.Windows.Forms.ColumnStyle(System.Windows.Forms.SizeType.Percent, 100F));
            this.tableLayoutPanel1.ColumnStyles.Add(new System.Windows.Forms.ColumnStyle());
            this.tableLayoutPanel1.ColumnStyles.Add(new System.Windows.Forms.ColumnStyle());
            this.tableLayoutPanel1.Controls.Add(this.ShortcutStartMenu, 2, 0);
            this.tableLayoutPanel1.Controls.Add(this.ShortcutLabel, 0, 0);
            this.tableLayoutPanel1.Controls.Add(this.ShortcutDesktop, 1, 0);
            this.tableLayoutPanel1.Location = new System.Drawing.Point(7, 255);
            this.tableLayoutPanel1.Name = "tableLayoutPanel1";
            this.tableLayoutPanel1.RowCount = 1;
            this.tableLayoutPanel1.RowStyles.Add(new System.Windows.Forms.RowStyle(System.Windows.Forms.SizeType.Percent, 100F));
            this.tableLayoutPanel1.Size = new System.Drawing.Size(274, 26);
            this.tableLayoutPanel1.TabIndex = 11;
            // 
            // ShortcutStartMenu
            // 
            this.ShortcutStartMenu.AutoSize = true;
            this.ShortcutStartMenu.Checked = true;
            this.ShortcutStartMenu.CheckState = System.Windows.Forms.CheckState.Checked;
            this.ShortcutStartMenu.Dock = System.Windows.Forms.DockStyle.Fill;
            this.ShortcutStartMenu.Location = new System.Drawing.Point(193, 3);
            this.ShortcutStartMenu.Name = "ShortcutStartMenu";
            this.ShortcutStartMenu.Padding = new System.Windows.Forms.Padding(0, 3, 0, 0);
            this.ShortcutStartMenu.Size = new System.Drawing.Size(78, 20);
            this.ShortcutStartMenu.TabIndex = 2;
            this.ShortcutStartMenu.Text = "Start Menu";
            this.ShortcutStartMenu.UseVisualStyleBackColor = true;
            // 
            // ShortcutLabel
            // 
            this.ShortcutLabel.AutoSize = true;
            this.ShortcutLabel.Dock = System.Windows.Forms.DockStyle.Fill;
            this.ShortcutLabel.Location = new System.Drawing.Point(3, 3);
            this.ShortcutLabel.Margin = new System.Windows.Forms.Padding(3);
            this.ShortcutLabel.Name = "ShortcutLabel";
            this.ShortcutLabel.Size = new System.Drawing.Size(112, 20);
            this.ShortcutLabel.TabIndex = 0;
            this.ShortcutLabel.Text = "Create shortcuts in:";
            this.ShortcutLabel.TextAlign = System.Drawing.ContentAlignment.MiddleCenter;
            // 
            // ShortcutDesktop
            // 
            this.ShortcutDesktop.AutoSize = true;
            this.ShortcutDesktop.Checked = true;
            this.ShortcutDesktop.CheckState = System.Windows.Forms.CheckState.Checked;
            this.ShortcutDesktop.Dock = System.Windows.Forms.DockStyle.Fill;
            this.ShortcutDesktop.Location = new System.Drawing.Point(121, 3);
            this.ShortcutDesktop.Name = "ShortcutDesktop";
            this.ShortcutDesktop.Padding = new System.Windows.Forms.Padding(0, 3, 0, 0);
            this.ShortcutDesktop.Size = new System.Drawing.Size(66, 20);
            this.ShortcutDesktop.TabIndex = 1;
            this.ShortcutDesktop.Text = "Desktop";
            this.ShortcutDesktop.UseVisualStyleBackColor = true;
            // 
            // TotalSizeDisplay
            // 
            this.TotalSizeDisplay.AutoSize = true;
            this.TotalSizeDisplay.Location = new System.Drawing.Point(478, 79);
            this.TotalSizeDisplay.Margin = new System.Windows.Forms.Padding(0);
            this.TotalSizeDisplay.Name = "TotalSizeDisplay";
            this.TotalSizeDisplay.Size = new System.Drawing.Size(0, 13);
            this.TotalSizeDisplay.TabIndex = 8;
            // 
            // TotalSizeLabel
            // 
            this.TotalSizeLabel.AutoSize = true;
            this.TotalSizeLabel.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Bold, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.TotalSizeLabel.Location = new System.Drawing.Point(410, 79);
            this.TotalSizeLabel.Margin = new System.Windows.Forms.Padding(0);
            this.TotalSizeLabel.Name = "TotalSizeLabel";
            this.TotalSizeLabel.Size = new System.Drawing.Size(68, 13);
            this.TotalSizeLabel.TabIndex = 7;
            this.TotalSizeLabel.Text = "Total Size:";
            // 
            // DownloadMessage
            // 
            this.DownloadMessage.Location = new System.Drawing.Point(410, 14);
            this.DownloadMessage.Name = "DownloadMessage";
            this.DownloadMessage.Size = new System.Drawing.Size(172, 30);
            this.DownloadMessage.TabIndex = 6;
            this.DownloadMessage.Text = "Choose components to include in your Flashpoint download.";
            // 
            // ComponentList
            // 
            this.ComponentList.Indent = 20;
            this.ComponentList.Location = new System.Drawing.Point(16, 14);
            this.ComponentList.Name = "ComponentList";
            this.ComponentList.Size = new System.Drawing.Size(384, 176);
            this.ComponentList.TabIndex = 5;
            this.ComponentList.TriStateStyleProperty = RikTheVeggie.TriStateTreeView.TriStateStyles.Installer;
            this.ComponentList.BeforeCheck += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeCheck);
            this.ComponentList.AfterCheck += new System.Windows.Forms.TreeViewEventHandler(this.ComponentList_AfterCheck);
            this.ComponentList.BeforeSelect += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeSelect);
            // 
            // ManageTab
            // 
            this.ManageTab.Controls.Add(this.ManagerMessage2);
            this.ManageTab.Controls.Add(this.ManagerMessage);
            this.ManageTab.Controls.Add(this.TotalSizeDisplay2);
            this.ManageTab.Controls.Add(this.TotalSizeLabel2);
            this.ManageTab.Controls.Add(this.UpdateButton);
            this.ManageTab.Controls.Add(this.ChangeButton);
            this.ManageTab.Controls.Add(this.SourcePathBox);
            this.ManageTab.Controls.Add(this.DescriptionBox2);
            this.ManageTab.Controls.Add(this.ComponentList2);
            this.ManageTab.Location = new System.Drawing.Point(4, 24);
            this.ManageTab.Name = "ManageTab";
            this.ManageTab.Padding = new System.Windows.Forms.Padding(3);
            this.ManageTab.Size = new System.Drawing.Size(592, 292);
            this.ManageTab.TabIndex = 1;
            this.ManageTab.Text = "Manage Components";
            this.ManageTab.UseVisualStyleBackColor = true;
            // 
            // ManagerMessage2
            // 
            this.ManagerMessage2.Location = new System.Drawing.Point(410, 44);
            this.ManagerMessage2.Name = "ManagerMessage2";
            this.ManagerMessage2.Size = new System.Drawing.Size(172, 30);
            this.ManagerMessage2.TabIndex = 21;
            this.ManagerMessage2.Text = "Select your Flashpoint folder to manage components.";
            // 
            // ManagerMessage
            // 
            this.ManagerMessage.Location = new System.Drawing.Point(410, 14);
            this.ManagerMessage.Name = "ManagerMessage";
            this.ManagerMessage.Size = new System.Drawing.Size(172, 30);
            this.ManagerMessage.TabIndex = 20;
            this.ManagerMessage.Text = "Choose components to add or remove from your Flashpoint copy.";
            // 
            // TotalSizeDisplay2
            // 
            this.TotalSizeDisplay2.AutoSize = true;
            this.TotalSizeDisplay2.Location = new System.Drawing.Point(492, 79);
            this.TotalSizeDisplay2.Margin = new System.Windows.Forms.Padding(0);
            this.TotalSizeDisplay2.Name = "TotalSizeDisplay2";
            this.TotalSizeDisplay2.Size = new System.Drawing.Size(0, 13);
            this.TotalSizeDisplay2.TabIndex = 19;
            this.TotalSizeDisplay2.Visible = false;
            // 
            // TotalSizeLabel2
            // 
            this.TotalSizeLabel2.AutoSize = true;
            this.TotalSizeLabel2.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Bold, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.TotalSizeLabel2.Location = new System.Drawing.Point(410, 79);
            this.TotalSizeLabel2.Margin = new System.Windows.Forms.Padding(0);
            this.TotalSizeLabel2.Name = "TotalSizeLabel2";
            this.TotalSizeLabel2.Size = new System.Drawing.Size(82, 13);
            this.TotalSizeLabel2.TabIndex = 18;
            this.TotalSizeLabel2.Text = "Size Change:";
            this.TotalSizeLabel2.Visible = false;
            // 
            // UpdateButton
            // 
            this.UpdateButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.UpdateButton.Enabled = false;
            this.UpdateButton.Location = new System.Drawing.Point(6, 256);
            this.UpdateButton.Name = "UpdateButton";
            this.UpdateButton.Size = new System.Drawing.Size(284, 26);
            this.UpdateButton.TabIndex = 17;
            this.UpdateButton.Text = "Check for updates";
            this.UpdateButton.UseVisualStyleBackColor = true;
            this.UpdateButton.Click += new System.EventHandler(this.UpdateButton_Click);
            // 
            // ChangeButton
            // 
            this.ChangeButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.ChangeButton.Enabled = false;
            this.ChangeButton.Location = new System.Drawing.Point(300, 256);
            this.ChangeButton.Name = "ChangeButton";
            this.ChangeButton.Size = new System.Drawing.Size(284, 26);
            this.ChangeButton.TabIndex = 16;
            this.ChangeButton.Text = "Apply changes";
            this.ChangeButton.UseVisualStyleBackColor = true;
            this.ChangeButton.Click += new System.EventHandler(this.ChangeButton_Click);
            // 
            // SourcePathBox
            // 
            this.SourcePathBox.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.SourcePathBox.Controls.Add(this.SourcePath);
            this.SourcePathBox.Controls.Add(this.SourcePathBrowse);
            this.SourcePathBox.Location = new System.Drawing.Point(7, 199);
            this.SourcePathBox.Name = "SourcePathBox";
            this.SourcePathBox.Size = new System.Drawing.Size(576, 49);
            this.SourcePathBox.TabIndex = 15;
            this.SourcePathBox.TabStop = false;
            this.SourcePathBox.Text = "Containing folder:";
            // 
            // SourcePath
            // 
            this.SourcePath.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Regular, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.SourcePath.Location = new System.Drawing.Point(9, 18);
            this.SourcePath.Name = "SourcePath";
            this.SourcePath.Size = new System.Drawing.Size(477, 20);
            this.SourcePath.TabIndex = 1;
            // 
            // SourcePathBrowse
            // 
            this.SourcePathBrowse.Location = new System.Drawing.Point(495, 17);
            this.SourcePathBrowse.Name = "SourcePathBrowse";
            this.SourcePathBrowse.Size = new System.Drawing.Size(73, 22);
            this.SourcePathBrowse.TabIndex = 0;
            this.SourcePathBrowse.Text = "Browse";
            this.SourcePathBrowse.UseVisualStyleBackColor = true;
            this.SourcePathBrowse.Click += new System.EventHandler(this.SourcePathBrowse_Click);
            // 
            // DescriptionBox2
            // 
            this.DescriptionBox2.Controls.Add(this.SizeDisplay2);
            this.DescriptionBox2.Controls.Add(this.SizeLabel2);
            this.DescriptionBox2.Controls.Add(this.Description2);
            this.DescriptionBox2.Location = new System.Drawing.Point(411, 100);
            this.DescriptionBox2.Name = "DescriptionBox2";
            this.DescriptionBox2.Padding = new System.Windows.Forms.Padding(10, 8, 10, 10);
            this.DescriptionBox2.Size = new System.Drawing.Size(163, 91);
            this.DescriptionBox2.TabIndex = 14;
            this.DescriptionBox2.TabStop = false;
            this.DescriptionBox2.Text = "Component Description";
            // 
            // SizeDisplay2
            // 
            this.SizeDisplay2.AutoSize = true;
            this.SizeDisplay2.Location = new System.Drawing.Point(45, 68);
            this.SizeDisplay2.Margin = new System.Windows.Forms.Padding(0);
            this.SizeDisplay2.Name = "SizeDisplay2";
            this.SizeDisplay2.Size = new System.Drawing.Size(38, 13);
            this.SizeDisplay2.TabIndex = 10;
            this.SizeDisplay2.Text = "0.0MB";
            this.SizeDisplay2.Visible = false;
            // 
            // SizeLabel2
            // 
            this.SizeLabel2.AutoSize = true;
            this.SizeLabel2.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Bold, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.SizeLabel2.Location = new System.Drawing.Point(10, 68);
            this.SizeLabel2.Margin = new System.Windows.Forms.Padding(0);
            this.SizeLabel2.Name = "SizeLabel2";
            this.SizeLabel2.Size = new System.Drawing.Size(35, 13);
            this.SizeLabel2.TabIndex = 9;
            this.SizeLabel2.Text = "Size:";
            this.SizeLabel2.Visible = false;
            // 
            // Description2
            // 
            this.Description2.Dock = System.Windows.Forms.DockStyle.Top;
            this.Description2.Location = new System.Drawing.Point(10, 21);
            this.Description2.Name = "Description2";
            this.Description2.Size = new System.Drawing.Size(143, 44);
            this.Description2.TabIndex = 0;
            // 
            // ComponentList2
            // 
            this.ComponentList2.Enabled = false;
            this.ComponentList2.Indent = 20;
            this.ComponentList2.Location = new System.Drawing.Point(16, 14);
            this.ComponentList2.Name = "ComponentList2";
            this.ComponentList2.Size = new System.Drawing.Size(384, 176);
            this.ComponentList2.TabIndex = 13;
            this.ComponentList2.TriStateStyleProperty = RikTheVeggie.TriStateTreeView.TriStateStyles.Installer;
            this.ComponentList2.BeforeSelect += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList2_BeforeSelect);
            // 
            // RemoveTab
            // 
            this.RemoveTab.Controls.Add(this.label1);
            this.RemoveTab.Controls.Add(this.RemoveShortcuts);
            this.RemoveTab.Controls.Add(this.SourcePathBox2);
            this.RemoveTab.Controls.Add(this.RemoveButton);
            this.RemoveTab.Location = new System.Drawing.Point(4, 24);
            this.RemoveTab.Name = "RemoveTab";
            this.RemoveTab.Padding = new System.Windows.Forms.Padding(3);
            this.RemoveTab.Size = new System.Drawing.Size(592, 292);
            this.RemoveTab.TabIndex = 2;
            this.RemoveTab.Text = "Remove Flashpoint";
            this.RemoveTab.UseVisualStyleBackColor = true;
            // 
            // label1
            // 
            this.label1.Location = new System.Drawing.Point(9, 11);
            this.label1.Name = "label1";
            this.label1.Size = new System.Drawing.Size(574, 40);
            this.label1.TabIndex = 20;
            this.label1.Text = "Remove a local Flashpoint copy by selecting its folder and then clicking the butt" +
    "on below.\r\n\r\nRemoving Flashpoint will not delete save data, but it will delete y" +
    "our favorites and any custom playlists.";
            this.label1.TextAlign = System.Drawing.ContentAlignment.MiddleCenter;
            // 
            // RemoveShortcuts
            // 
            this.RemoveShortcuts.AutoSize = true;
            this.RemoveShortcuts.Checked = true;
            this.RemoveShortcuts.CheckState = System.Windows.Forms.CheckState.Checked;
            this.RemoveShortcuts.Location = new System.Drawing.Point(169, 72);
            this.RemoveShortcuts.Name = "RemoveShortcuts";
            this.RemoveShortcuts.Size = new System.Drawing.Size(239, 17);
            this.RemoveShortcuts.TabIndex = 19;
            this.RemoveShortcuts.Text = "Remove shortcuts in desktop and Start menu";
            this.RemoveShortcuts.UseVisualStyleBackColor = true;
            // 
            // SourcePathBox2
            // 
            this.SourcePathBox2.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.SourcePathBox2.Controls.Add(this.SourcePath2);
            this.SourcePathBox2.Controls.Add(this.SourcePathBrowse2);
            this.SourcePathBox2.Location = new System.Drawing.Point(7, 199);
            this.SourcePathBox2.Name = "SourcePathBox2";
            this.SourcePathBox2.Size = new System.Drawing.Size(576, 49);
            this.SourcePathBox2.TabIndex = 18;
            this.SourcePathBox2.TabStop = false;
            this.SourcePathBox2.Text = "Containing folder:";
            // 
            // SourcePath2
            // 
            this.SourcePath2.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Regular, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.SourcePath2.Location = new System.Drawing.Point(9, 18);
            this.SourcePath2.Name = "SourcePath2";
            this.SourcePath2.Size = new System.Drawing.Size(477, 20);
            this.SourcePath2.TabIndex = 1;
            // 
            // SourcePathBrowse2
            // 
            this.SourcePathBrowse2.Location = new System.Drawing.Point(495, 17);
            this.SourcePathBrowse2.Name = "SourcePathBrowse2";
            this.SourcePathBrowse2.Size = new System.Drawing.Size(73, 22);
            this.SourcePathBrowse2.TabIndex = 0;
            this.SourcePathBrowse2.Text = "Browse";
            this.SourcePathBrowse2.UseVisualStyleBackColor = true;
            this.SourcePathBrowse2.Click += new System.EventHandler(this.SourcePathBrowse2_Click);
            // 
            // RemoveButton
            // 
            this.RemoveButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.RemoveButton.Enabled = false;
            this.RemoveButton.Location = new System.Drawing.Point(6, 256);
            this.RemoveButton.Name = "RemoveButton";
            this.RemoveButton.Size = new System.Drawing.Size(578, 26);
            this.RemoveButton.TabIndex = 17;
            this.RemoveButton.Text = "Remove Flashpoint";
            this.RemoveButton.UseVisualStyleBackColor = true;
            this.RemoveButton.Click += new System.EventHandler(this.RemoveButton_Click);
            // 
            // Main
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(624, 561);
            this.Controls.Add(this.TabControl);
            this.Controls.Add(this.Link);
            this.Controls.Add(this.About);
            this.Controls.Add(this.Logo);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.Icon = ((System.Drawing.Icon)(resources.GetObject("$this.Icon")));
            this.MaximizeBox = false;
            this.MinimizeBox = false;
            this.Name = "Main";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "Flashpoint Manager";
            this.Load += new System.EventHandler(this.Main_Load);
            ((System.ComponentModel.ISupportInitialize)(this.Logo)).EndInit();
            this.DestinationPathBox.ResumeLayout(false);
            this.DestinationPathBox.PerformLayout();
            this.TabControl.ResumeLayout(false);
            this.DownloadTab.ResumeLayout(false);
            this.DownloadTab.PerformLayout();
            this.DescriptionBox.ResumeLayout(false);
            this.DescriptionBox.PerformLayout();
            this.tableLayoutPanel1.ResumeLayout(false);
            this.tableLayoutPanel1.PerformLayout();
            this.ManageTab.ResumeLayout(false);
            this.ManageTab.PerformLayout();
            this.SourcePathBox.ResumeLayout(false);
            this.SourcePathBox.PerformLayout();
            this.DescriptionBox2.ResumeLayout(false);
            this.DescriptionBox2.PerformLayout();
            this.RemoveTab.ResumeLayout(false);
            this.RemoveTab.PerformLayout();
            this.SourcePathBox2.ResumeLayout(false);
            this.SourcePathBox2.PerformLayout();
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private System.Windows.Forms.PictureBox Logo;
        private System.Windows.Forms.Label About;
        private System.Windows.Forms.GroupBox DestinationPathBox;
        public System.Windows.Forms.TextBox DestinationPath;
        private System.Windows.Forms.Button DestinationPathBrowse;
        private System.Windows.Forms.Button DownloadButton;
        private System.Windows.Forms.LinkLabel Link;
        private System.Windows.Forms.TabControl TabControl;
        private System.Windows.Forms.TabPage DownloadTab;
        private System.Windows.Forms.TabPage ManageTab;
        private System.Windows.Forms.Label TotalSizeLabel;
        private System.Windows.Forms.Label DownloadMessage;
        public RikTheVeggie.TriStateTreeView ComponentList;
        private System.Windows.Forms.TabPage RemoveTab;
        private System.Windows.Forms.TableLayoutPanel tableLayoutPanel1;
        private System.Windows.Forms.Label ShortcutLabel;
        public System.Windows.Forms.CheckBox ShortcutStartMenu;
        public System.Windows.Forms.CheckBox ShortcutDesktop;
        public RikTheVeggie.TriStateTreeView ComponentList2;
        private System.Windows.Forms.GroupBox DescriptionBox;
        private System.Windows.Forms.Label Description;
        private System.Windows.Forms.Label DownloadMessage2;
        private System.Windows.Forms.Button UpdateButton;
        private System.Windows.Forms.GroupBox SourcePathBox;
        public System.Windows.Forms.TextBox SourcePath;
        private System.Windows.Forms.Button SourcePathBrowse;
        private System.Windows.Forms.GroupBox DescriptionBox2;
        private System.Windows.Forms.Label Description2;
        private System.Windows.Forms.Label TotalSizeLabel2;
        private System.Windows.Forms.Label ManagerMessage;
        private System.Windows.Forms.Label ManagerMessage2;
        public System.Windows.Forms.Label TotalSizeDisplay;
        public System.Windows.Forms.Label TotalSizeDisplay2;
        public System.Windows.Forms.Button ChangeButton;
        private System.Windows.Forms.Label label1;
        private System.Windows.Forms.GroupBox SourcePathBox2;
        public System.Windows.Forms.TextBox SourcePath2;
        private System.Windows.Forms.Button SourcePathBrowse2;
        public System.Windows.Forms.Button RemoveButton;
        public System.Windows.Forms.CheckBox RemoveShortcuts;
        public System.Windows.Forms.Label SizeDisplay;
        private System.Windows.Forms.Label SizeLabel;
        public System.Windows.Forms.Label SizeDisplay2;
        private System.Windows.Forms.Label SizeLabel2;
    }
}

