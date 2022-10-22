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
            this.Folder = new System.Windows.Forms.GroupBox();
            this.FolderTextBox = new System.Windows.Forms.TextBox();
            this.FolderButton = new System.Windows.Forms.Button();
            this.Install = new System.Windows.Forms.Button();
            this.Link = new System.Windows.Forms.LinkLabel();
            this.TabControl = new System.Windows.Forms.TabControl();
            this.DownloadTab = new System.Windows.Forms.TabPage();
            this.ComponentMessage2 = new System.Windows.Forms.Label();
            this.DescriptionBox = new System.Windows.Forms.GroupBox();
            this.Description = new System.Windows.Forms.Label();
            this.tableLayoutPanel1 = new System.Windows.Forms.TableLayoutPanel();
            this.ShortcutStartMenu = new System.Windows.Forms.CheckBox();
            this.ShortcutLabel = new System.Windows.Forms.Label();
            this.ShortcutDesktop = new System.Windows.Forms.CheckBox();
            this.ComponentSize = new System.Windows.Forms.Label();
            this.ComponentSizeLabel = new System.Windows.Forms.Label();
            this.ComponentMessage = new System.Windows.Forms.Label();
            this.ComponentQueue = new RikTheVeggie.TriStateTreeView();
            this.UpdateTab = new System.Windows.Forms.TabPage();
            this.ComponentManager = new RikTheVeggie.TriStateTreeView();
            this.RemoveTab = new System.Windows.Forms.TabPage();
            ((System.ComponentModel.ISupportInitialize)(this.Logo)).BeginInit();
            this.Folder.SuspendLayout();
            this.TabControl.SuspendLayout();
            this.DownloadTab.SuspendLayout();
            this.DescriptionBox.SuspendLayout();
            this.tableLayoutPanel1.SuspendLayout();
            this.UpdateTab.SuspendLayout();
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
            this.About.Size = new System.Drawing.Size(124, 13);
            this.About.TabIndex = 1;
            this.About.Text = "Flashpoint Manager v1.0";
            // 
            // Folder
            // 
            this.Folder.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.Folder.Controls.Add(this.FolderTextBox);
            this.Folder.Controls.Add(this.FolderButton);
            this.Folder.Location = new System.Drawing.Point(7, 199);
            this.Folder.Name = "Folder";
            this.Folder.Size = new System.Drawing.Size(576, 49);
            this.Folder.TabIndex = 2;
            this.Folder.TabStop = false;
            this.Folder.Text = "Destination folder:";
            // 
            // FolderTextBox
            // 
            this.FolderTextBox.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Regular, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.FolderTextBox.Location = new System.Drawing.Point(9, 18);
            this.FolderTextBox.Name = "FolderTextBox";
            this.FolderTextBox.Size = new System.Drawing.Size(477, 20);
            this.FolderTextBox.TabIndex = 1;
            // 
            // FolderButton
            // 
            this.FolderButton.Location = new System.Drawing.Point(495, 17);
            this.FolderButton.Name = "FolderButton";
            this.FolderButton.Size = new System.Drawing.Size(73, 22);
            this.FolderButton.TabIndex = 0;
            this.FolderButton.Text = "Browse";
            this.FolderButton.UseVisualStyleBackColor = true;
            this.FolderButton.Click += new System.EventHandler(this.FolderButton_Click);
            // 
            // Install
            // 
            this.Install.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.Install.Location = new System.Drawing.Point(300, 256);
            this.Install.Name = "Install";
            this.Install.Size = new System.Drawing.Size(284, 26);
            this.Install.TabIndex = 2;
            this.Install.Text = "Download Flashpoint";
            this.Install.UseVisualStyleBackColor = true;
            this.Install.Click += new System.EventHandler(this.Install_Click);
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
            this.TabControl.Controls.Add(this.UpdateTab);
            this.TabControl.Controls.Add(this.RemoveTab);
            this.TabControl.ItemSize = new System.Drawing.Size(150, 20);
            this.TabControl.Location = new System.Drawing.Point(12, 229);
            this.TabControl.Name = "TabControl";
            this.TabControl.SelectedIndex = 0;
            this.TabControl.Size = new System.Drawing.Size(600, 320);
            this.TabControl.SizeMode = System.Windows.Forms.TabSizeMode.Fixed;
            this.TabControl.TabIndex = 5;
            // 
            // DownloadTab
            // 
            this.DownloadTab.Controls.Add(this.ComponentMessage2);
            this.DownloadTab.Controls.Add(this.DescriptionBox);
            this.DownloadTab.Controls.Add(this.tableLayoutPanel1);
            this.DownloadTab.Controls.Add(this.ComponentSize);
            this.DownloadTab.Controls.Add(this.ComponentSizeLabel);
            this.DownloadTab.Controls.Add(this.ComponentMessage);
            this.DownloadTab.Controls.Add(this.ComponentQueue);
            this.DownloadTab.Controls.Add(this.Folder);
            this.DownloadTab.Controls.Add(this.Install);
            this.DownloadTab.Location = new System.Drawing.Point(4, 24);
            this.DownloadTab.Name = "DownloadTab";
            this.DownloadTab.Padding = new System.Windows.Forms.Padding(3);
            this.DownloadTab.Size = new System.Drawing.Size(592, 292);
            this.DownloadTab.TabIndex = 0;
            this.DownloadTab.Text = "Download Flashpoint";
            this.DownloadTab.UseVisualStyleBackColor = true;
            // 
            // ComponentMessage2
            // 
            this.ComponentMessage2.Location = new System.Drawing.Point(410, 44);
            this.ComponentMessage2.Name = "ComponentMessage2";
            this.ComponentMessage2.Size = new System.Drawing.Size(172, 30);
            this.ComponentMessage2.TabIndex = 13;
            this.ComponentMessage2.Text = "Click on a component to learn more about it.";
            // 
            // DescriptionBox
            // 
            this.DescriptionBox.Controls.Add(this.Description);
            this.DescriptionBox.Location = new System.Drawing.Point(411, 116);
            this.DescriptionBox.Name = "DescriptionBox";
            this.DescriptionBox.Padding = new System.Windows.Forms.Padding(10, 8, 10, 10);
            this.DescriptionBox.Size = new System.Drawing.Size(163, 75);
            this.DescriptionBox.TabIndex = 12;
            this.DescriptionBox.TabStop = false;
            this.DescriptionBox.Text = "Component Description";
            // 
            // Description
            // 
            this.Description.Dock = System.Windows.Forms.DockStyle.Fill;
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
            // ComponentSize
            // 
            this.ComponentSize.Location = new System.Drawing.Point(477, 85);
            this.ComponentSize.Name = "ComponentSize";
            this.ComponentSize.Size = new System.Drawing.Size(105, 15);
            this.ComponentSize.TabIndex = 8;
            this.ComponentSize.Text = "0MB";
            // 
            // ComponentSizeLabel
            // 
            this.ComponentSizeLabel.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Bold, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.ComponentSizeLabel.Location = new System.Drawing.Point(410, 85);
            this.ComponentSizeLabel.Name = "ComponentSizeLabel";
            this.ComponentSizeLabel.Size = new System.Drawing.Size(70, 15);
            this.ComponentSizeLabel.TabIndex = 7;
            this.ComponentSizeLabel.Text = "Total Size:";
            // 
            // ComponentMessage
            // 
            this.ComponentMessage.Location = new System.Drawing.Point(410, 14);
            this.ComponentMessage.Name = "ComponentMessage";
            this.ComponentMessage.Size = new System.Drawing.Size(172, 30);
            this.ComponentMessage.TabIndex = 6;
            this.ComponentMessage.Text = "Choose components to include in your Flashpoint download.";
            // 
            // ComponentQueue
            // 
            this.ComponentQueue.Indent = 20;
            this.ComponentQueue.Location = new System.Drawing.Point(16, 14);
            this.ComponentQueue.Name = "ComponentQueue";
            this.ComponentQueue.Size = new System.Drawing.Size(384, 176);
            this.ComponentQueue.TabIndex = 5;
            this.ComponentQueue.TriStateStyleProperty = RikTheVeggie.TriStateTreeView.TriStateStyles.Installer;
            this.ComponentQueue.BeforeCheck += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeCheck);
            this.ComponentQueue.AfterCheck += new System.Windows.Forms.TreeViewEventHandler(this.ComponentList_AfterCheck);
            this.ComponentQueue.BeforeSelect += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentQueue_BeforeSelect);
            // 
            // UpdateTab
            // 
            this.UpdateTab.Controls.Add(this.ComponentManager);
            this.UpdateTab.Location = new System.Drawing.Point(4, 24);
            this.UpdateTab.Name = "UpdateTab";
            this.UpdateTab.Padding = new System.Windows.Forms.Padding(3);
            this.UpdateTab.Size = new System.Drawing.Size(592, 292);
            this.UpdateTab.TabIndex = 1;
            this.UpdateTab.Text = "Manage Components";
            this.UpdateTab.UseVisualStyleBackColor = true;
            // 
            // ComponentManager
            // 
            this.ComponentManager.Indent = 20;
            this.ComponentManager.Location = new System.Drawing.Point(16, 14);
            this.ComponentManager.Name = "ComponentManager";
            this.ComponentManager.Size = new System.Drawing.Size(384, 176);
            this.ComponentManager.TabIndex = 13;
            this.ComponentManager.TriStateStyleProperty = RikTheVeggie.TriStateTreeView.TriStateStyles.Installer;
            this.ComponentManager.BeforeCheck += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeCheck);
            // 
            // RemoveTab
            // 
            this.RemoveTab.Location = new System.Drawing.Point(4, 24);
            this.RemoveTab.Name = "RemoveTab";
            this.RemoveTab.Padding = new System.Windows.Forms.Padding(3);
            this.RemoveTab.Size = new System.Drawing.Size(592, 292);
            this.RemoveTab.TabIndex = 2;
            this.RemoveTab.Text = "Remove Flashpoint";
            this.RemoveTab.UseVisualStyleBackColor = true;
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
            this.Folder.ResumeLayout(false);
            this.Folder.PerformLayout();
            this.TabControl.ResumeLayout(false);
            this.DownloadTab.ResumeLayout(false);
            this.DescriptionBox.ResumeLayout(false);
            this.tableLayoutPanel1.ResumeLayout(false);
            this.tableLayoutPanel1.PerformLayout();
            this.UpdateTab.ResumeLayout(false);
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private System.Windows.Forms.PictureBox Logo;
        private System.Windows.Forms.Label About;
        private System.Windows.Forms.GroupBox Folder;
        public System.Windows.Forms.TextBox FolderTextBox;
        private System.Windows.Forms.Button FolderButton;
        private System.Windows.Forms.Button Install;
        private System.Windows.Forms.LinkLabel Link;
        private System.Windows.Forms.TabControl TabControl;
        private System.Windows.Forms.TabPage DownloadTab;
        private System.Windows.Forms.TabPage UpdateTab;
        private System.Windows.Forms.Label ComponentSize;
        private System.Windows.Forms.Label ComponentSizeLabel;
        private System.Windows.Forms.Label ComponentMessage;
        public RikTheVeggie.TriStateTreeView ComponentQueue;
        private System.Windows.Forms.TabPage RemoveTab;
        private System.Windows.Forms.TableLayoutPanel tableLayoutPanel1;
        private System.Windows.Forms.Label ShortcutLabel;
        public System.Windows.Forms.CheckBox ShortcutStartMenu;
        public System.Windows.Forms.CheckBox ShortcutDesktop;
        public RikTheVeggie.TriStateTreeView ComponentManager;
        private System.Windows.Forms.GroupBox DescriptionBox;
        private System.Windows.Forms.Label Description;
        private System.Windows.Forms.Label ComponentMessage2;
    }
}

