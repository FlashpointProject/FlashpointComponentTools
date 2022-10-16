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
            this.Shortcut = new System.Windows.Forms.CheckBox();
            this.Link = new System.Windows.Forms.LinkLabel();
            this.TabControl = new System.Windows.Forms.TabControl();
            this.InstallTab = new System.Windows.Forms.TabPage();
            this.UpdateTab = new System.Windows.Forms.TabPage();
            this.UninstallTab = new System.Windows.Forms.TabPage();
            this.ComponentList = new RikTheVeggie.TriStateTreeView();
            this.ComponentMessage = new System.Windows.Forms.Label();
            this.ComponentSizeLabel = new System.Windows.Forms.Label();
            this.ComponentSize = new System.Windows.Forms.Label();
            ((System.ComponentModel.ISupportInitialize)(this.Logo)).BeginInit();
            this.Folder.SuspendLayout();
            this.TabControl.SuspendLayout();
            this.InstallTab.SuspendLayout();
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
            this.About.Size = new System.Drawing.Size(118, 13);
            this.About.TabIndex = 1;
            this.About.Text = "Flashpoint Installer v1.0";
            // 
            // Folder
            // 
            this.Folder.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.Folder.Controls.Add(this.FolderTextBox);
            this.Folder.Controls.Add(this.FolderButton);
            this.Folder.Location = new System.Drawing.Point(7, 239);
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
            this.FolderTextBox.Size = new System.Drawing.Size(473, 20);
            this.FolderTextBox.TabIndex = 1;
            // 
            // FolderButton
            // 
            this.FolderButton.Location = new System.Drawing.Point(491, 17);
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
            this.Install.Location = new System.Drawing.Point(168, 296);
            this.Install.Name = "Install";
            this.Install.Size = new System.Drawing.Size(416, 26);
            this.Install.TabIndex = 2;
            this.Install.Text = "Install Flashpoint";
            this.Install.UseVisualStyleBackColor = true;
            this.Install.Click += new System.EventHandler(this.Install_Click);
            // 
            // Shortcut
            // 
            this.Shortcut.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.Shortcut.AutoSize = true;
            this.Shortcut.Checked = true;
            this.Shortcut.CheckState = System.Windows.Forms.CheckState.Checked;
            this.Shortcut.Location = new System.Drawing.Point(7, 302);
            this.Shortcut.Name = "Shortcut";
            this.Shortcut.Size = new System.Drawing.Size(154, 17);
            this.Shortcut.TabIndex = 3;
            this.Shortcut.Text = "Create shortcut on desktop";
            this.Shortcut.UseVisualStyleBackColor = true;
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
            this.TabControl.Controls.Add(this.InstallTab);
            this.TabControl.Controls.Add(this.UpdateTab);
            this.TabControl.Controls.Add(this.UninstallTab);
            this.TabControl.ItemSize = new System.Drawing.Size(150, 20);
            this.TabControl.Location = new System.Drawing.Point(12, 229);
            this.TabControl.Name = "TabControl";
            this.TabControl.SelectedIndex = 0;
            this.TabControl.Size = new System.Drawing.Size(600, 360);
            this.TabControl.SizeMode = System.Windows.Forms.TabSizeMode.Fixed;
            this.TabControl.TabIndex = 5;
            // 
            // InstallTab
            // 
            this.InstallTab.Controls.Add(this.ComponentSize);
            this.InstallTab.Controls.Add(this.ComponentSizeLabel);
            this.InstallTab.Controls.Add(this.ComponentMessage);
            this.InstallTab.Controls.Add(this.ComponentList);
            this.InstallTab.Controls.Add(this.Folder);
            this.InstallTab.Controls.Add(this.Install);
            this.InstallTab.Controls.Add(this.Shortcut);
            this.InstallTab.Location = new System.Drawing.Point(4, 24);
            this.InstallTab.Name = "InstallTab";
            this.InstallTab.Padding = new System.Windows.Forms.Padding(3);
            this.InstallTab.Size = new System.Drawing.Size(592, 332);
            this.InstallTab.TabIndex = 0;
            this.InstallTab.Text = "Install";
            this.InstallTab.UseVisualStyleBackColor = true;
            // 
            // UpdateTab
            // 
            this.UpdateTab.Location = new System.Drawing.Point(4, 24);
            this.UpdateTab.Name = "UpdateTab";
            this.UpdateTab.Padding = new System.Windows.Forms.Padding(3);
            this.UpdateTab.Size = new System.Drawing.Size(592, 332);
            this.UpdateTab.TabIndex = 1;
            this.UpdateTab.Text = "Update";
            this.UpdateTab.UseVisualStyleBackColor = true;
            // 
            // UninstallTab
            // 
            this.UninstallTab.Location = new System.Drawing.Point(4, 24);
            this.UninstallTab.Name = "UninstallTab";
            this.UninstallTab.Padding = new System.Windows.Forms.Padding(3);
            this.UninstallTab.Size = new System.Drawing.Size(592, 332);
            this.UninstallTab.TabIndex = 2;
            this.UninstallTab.Text = "Uninstall";
            this.UninstallTab.UseVisualStyleBackColor = true;
            // 
            // ComponentList
            // 
            this.ComponentList.Indent = 20;
            this.ComponentList.Location = new System.Drawing.Point(16, 14);
            this.ComponentList.Name = "ComponentList";
            this.ComponentList.Size = new System.Drawing.Size(384, 216);
            this.ComponentList.TabIndex = 5;
            this.ComponentList.TriStateStyleProperty = RikTheVeggie.TriStateTreeView.TriStateStyles.Installer;
            this.ComponentList.BeforeCheck += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeCheck);
            this.ComponentList.AfterCheck += new System.Windows.Forms.TreeViewEventHandler(this.ComponentList_AfterCheck);
            // 
            // ComponentMessage
            // 
            this.ComponentMessage.Location = new System.Drawing.Point(410, 14);
            this.ComponentMessage.Name = "ComponentMessage";
            this.ComponentMessage.Size = new System.Drawing.Size(172, 30);
            this.ComponentMessage.TabIndex = 6;
            this.ComponentMessage.Text = "Select components to be included in your Flashpoint installation.";
            // 
            // ComponentSizeLabel
            // 
            this.ComponentSizeLabel.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Bold, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.ComponentSizeLabel.Location = new System.Drawing.Point(410, 59);
            this.ComponentSizeLabel.Name = "ComponentSizeLabel";
            this.ComponentSizeLabel.Size = new System.Drawing.Size(40, 15);
            this.ComponentSizeLabel.TabIndex = 7;
            this.ComponentSizeLabel.Text = "Size:";
            // 
            // ComponentSize
            // 
            this.ComponentSize.Location = new System.Drawing.Point(456, 59);
            this.ComponentSize.Name = "ComponentSize";
            this.ComponentSize.Size = new System.Drawing.Size(126, 15);
            this.ComponentSize.TabIndex = 8;
            // 
            // Main
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(624, 601);
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
            this.Text = "Flashpoint Installer";
            this.Load += new System.EventHandler(this.Main_Load);
            ((System.ComponentModel.ISupportInitialize)(this.Logo)).EndInit();
            this.Folder.ResumeLayout(false);
            this.Folder.PerformLayout();
            this.TabControl.ResumeLayout(false);
            this.InstallTab.ResumeLayout(false);
            this.InstallTab.PerformLayout();
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
        public System.Windows.Forms.CheckBox Shortcut;
        private System.Windows.Forms.TabControl TabControl;
        private System.Windows.Forms.TabPage InstallTab;
        private System.Windows.Forms.TabPage UpdateTab;
        private System.Windows.Forms.TabPage UninstallTab;
        private RikTheVeggie.TriStateTreeView ComponentList;
        private System.Windows.Forms.Label ComponentSize;
        private System.Windows.Forms.Label ComponentSizeLabel;
        private System.Windows.Forms.Label ComponentMessage;
    }
}

