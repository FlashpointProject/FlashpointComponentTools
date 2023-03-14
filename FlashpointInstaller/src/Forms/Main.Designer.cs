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
            this.InstallButton = new System.Windows.Forms.Button();
            this.Link = new System.Windows.Forms.LinkLabel();
            this.Message2 = new System.Windows.Forms.Label();
            this.DescriptionBox = new System.Windows.Forms.GroupBox();
            this.Description = new System.Windows.Forms.Label();
            this.ShortcutTable = new System.Windows.Forms.TableLayoutPanel();
            this.ShortcutStartMenu = new System.Windows.Forms.CheckBox();
            this.ShortcutDesktop = new System.Windows.Forms.CheckBox();
            this.ShortcutLabel = new System.Windows.Forms.Label();
            this.Message = new System.Windows.Forms.Label();
            this.ComponentList = new RikTheVeggie.TriStateTreeView();
            ((System.ComponentModel.ISupportInitialize)(this.Logo)).BeginInit();
            this.DestinationPathBox.SuspendLayout();
            this.DescriptionBox.SuspendLayout();
            this.ShortcutTable.SuspendLayout();
            this.SuspendLayout();
            // 
            // Logo
            // 
            this.Logo.Image = ((System.Drawing.Image)(resources.GetObject("Logo.Image")));
            this.Logo.Location = new System.Drawing.Point(21, 12);
            this.Logo.Name = "Logo";
            this.Logo.Size = new System.Drawing.Size(591, 177);
            this.Logo.SizeMode = System.Windows.Forms.PictureBoxSizeMode.CenterImage;
            this.Logo.TabIndex = 0;
            this.Logo.TabStop = false;
            // 
            // About
            // 
            this.About.AutoSize = true;
            this.About.Location = new System.Drawing.Point(18, 192);
            this.About.Name = "About";
            this.About.Size = new System.Drawing.Size(94, 13);
            this.About.TabIndex = 1;
            this.About.Text = "Flashpoint Installer";
            // 
            // DestinationPathBox
            // 
            this.DestinationPathBox.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.DestinationPathBox.Controls.Add(this.DestinationPath);
            this.DestinationPathBox.Controls.Add(this.DestinationPathBrowse);
            this.DestinationPathBox.Location = new System.Drawing.Point(12, 384);
            this.DestinationPathBox.Name = "DestinationPathBox";
            this.DestinationPathBox.Size = new System.Drawing.Size(600, 49);
            this.DestinationPathBox.TabIndex = 2;
            this.DestinationPathBox.TabStop = false;
            this.DestinationPathBox.Text = "Destination Folder";
            // 
            // DestinationPath
            // 
            this.DestinationPath.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Regular, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.DestinationPath.Location = new System.Drawing.Point(9, 18);
            this.DestinationPath.Name = "DestinationPath";
            this.DestinationPath.Size = new System.Drawing.Size(503, 20);
            this.DestinationPath.TabIndex = 1;
            this.DestinationPath.KeyPress += new System.Windows.Forms.KeyPressEventHandler(this.DestinationPath_KeyPress);
            // 
            // DestinationPathBrowse
            // 
            this.DestinationPathBrowse.Location = new System.Drawing.Point(519, 17);
            this.DestinationPathBrowse.Name = "DestinationPathBrowse";
            this.DestinationPathBrowse.Size = new System.Drawing.Size(73, 22);
            this.DestinationPathBrowse.TabIndex = 0;
            this.DestinationPathBrowse.Text = "Browse";
            this.DestinationPathBrowse.UseVisualStyleBackColor = true;
            this.DestinationPathBrowse.Click += new System.EventHandler(this.DestinationPathBrowse_Click);
            // 
            // InstallButton
            // 
            this.InstallButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.InstallButton.Location = new System.Drawing.Point(300, 443);
            this.InstallButton.Name = "InstallButton";
            this.InstallButton.Size = new System.Drawing.Size(313, 26);
            this.InstallButton.TabIndex = 2;
            this.InstallButton.Text = "Install Flashpoint";
            this.InstallButton.UseVisualStyleBackColor = true;
            this.InstallButton.Click += new System.EventHandler(this.InstallButton_Click);
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
            this.Link.TextAlign = System.Drawing.ContentAlignment.TopRight;
            this.Link.LinkClicked += new System.Windows.Forms.LinkLabelLinkClickedEventHandler(this.Link_LinkClicked);
            // 
            // Message2
            // 
            this.Message2.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.Message2.Location = new System.Drawing.Point(415, 257);
            this.Message2.Name = "Message2";
            this.Message2.Size = new System.Drawing.Size(197, 30);
            this.Message2.TabIndex = 13;
            this.Message2.Text = "Click on a component or category to learn more about it.";
            // 
            // DescriptionBox
            // 
            this.DescriptionBox.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.DescriptionBox.Controls.Add(this.Description);
            this.DescriptionBox.Location = new System.Drawing.Point(415, 305);
            this.DescriptionBox.Name = "DescriptionBox";
            this.DescriptionBox.Padding = new System.Windows.Forms.Padding(10, 8, 10, 10);
            this.DescriptionBox.Size = new System.Drawing.Size(197, 72);
            this.DescriptionBox.TabIndex = 12;
            this.DescriptionBox.TabStop = false;
            this.DescriptionBox.Text = "Component Description";
            this.DescriptionBox.Visible = false;
            // 
            // Description
            // 
            this.Description.Location = new System.Drawing.Point(10, 21);
            this.Description.Name = "Description";
            this.Description.Size = new System.Drawing.Size(177, 41);
            this.Description.TabIndex = 0;
            // 
            // ShortcutTable
            // 
            this.ShortcutTable.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.ShortcutTable.ColumnCount = 3;
            this.ShortcutTable.ColumnStyles.Add(new System.Windows.Forms.ColumnStyle(System.Windows.Forms.SizeType.Percent, 100F));
            this.ShortcutTable.ColumnStyles.Add(new System.Windows.Forms.ColumnStyle());
            this.ShortcutTable.ColumnStyles.Add(new System.Windows.Forms.ColumnStyle());
            this.ShortcutTable.Controls.Add(this.ShortcutStartMenu, 2, 0);
            this.ShortcutTable.Controls.Add(this.ShortcutDesktop, 1, 0);
            this.ShortcutTable.Controls.Add(this.ShortcutLabel, 0, 0);
            this.ShortcutTable.Location = new System.Drawing.Point(12, 442);
            this.ShortcutTable.Name = "ShortcutTable";
            this.ShortcutTable.RowCount = 1;
            this.ShortcutTable.RowStyles.Add(new System.Windows.Forms.RowStyle(System.Windows.Forms.SizeType.Percent, 100F));
            this.ShortcutTable.Size = new System.Drawing.Size(274, 26);
            this.ShortcutTable.TabIndex = 11;
            // 
            // ShortcutStartMenu
            // 
            this.ShortcutStartMenu.AutoSize = true;
            this.ShortcutStartMenu.Checked = true;
            this.ShortcutStartMenu.CheckState = System.Windows.Forms.CheckState.Checked;
            this.ShortcutStartMenu.Location = new System.Drawing.Point(193, 3);
            this.ShortcutStartMenu.Name = "ShortcutStartMenu";
            this.ShortcutStartMenu.Padding = new System.Windows.Forms.Padding(0, 3, 0, 0);
            this.ShortcutStartMenu.Size = new System.Drawing.Size(78, 20);
            this.ShortcutStartMenu.TabIndex = 2;
            this.ShortcutStartMenu.Text = "Start Menu";
            this.ShortcutStartMenu.UseVisualStyleBackColor = true;
            // 
            // ShortcutDesktop
            // 
            this.ShortcutDesktop.AutoSize = true;
            this.ShortcutDesktop.Checked = true;
            this.ShortcutDesktop.CheckState = System.Windows.Forms.CheckState.Checked;
            this.ShortcutDesktop.Location = new System.Drawing.Point(121, 3);
            this.ShortcutDesktop.Name = "ShortcutDesktop";
            this.ShortcutDesktop.Padding = new System.Windows.Forms.Padding(0, 3, 0, 0);
            this.ShortcutDesktop.Size = new System.Drawing.Size(66, 20);
            this.ShortcutDesktop.TabIndex = 1;
            this.ShortcutDesktop.Text = "Desktop";
            this.ShortcutDesktop.UseVisualStyleBackColor = true;
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
            // Message
            // 
            this.Message.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.Message.Location = new System.Drawing.Point(415, 225);
            this.Message.Name = "Message";
            this.Message.Size = new System.Drawing.Size(197, 30);
            this.Message.TabIndex = 6;
            this.Message.Text = "Choose components to include in your Flashpoint installation.";
            // 
            // ComponentList
            // 
            this.ComponentList.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.ComponentList.Indent = 20;
            this.ComponentList.Location = new System.Drawing.Point(21, 220);
            this.ComponentList.Name = "ComponentList";
            this.ComponentList.Size = new System.Drawing.Size(384, 156);
            this.ComponentList.TabIndex = 5;
            this.ComponentList.TriStateStyleProperty = RikTheVeggie.TriStateTreeView.TriStateStyles.Installer;
            this.ComponentList.BeforeCheck += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeCheck);
            this.ComponentList.AfterCheck += new System.Windows.Forms.TreeViewEventHandler(this.ComponentList_AfterCheck);
            this.ComponentList.BeforeSelect += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeSelect);
            // 
            // Main
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(624, 481);
            this.Controls.Add(this.DescriptionBox);
            this.Controls.Add(this.ShortcutTable);
            this.Controls.Add(this.Message);
            this.Controls.Add(this.ComponentList);
            this.Controls.Add(this.DestinationPathBox);
            this.Controls.Add(this.InstallButton);
            this.Controls.Add(this.Link);
            this.Controls.Add(this.Message2);
            this.Controls.Add(this.About);
            this.Controls.Add(this.Logo);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.Icon = ((System.Drawing.Icon)(resources.GetObject("$this.Icon")));
            this.MaximizeBox = false;
            this.Name = "Main";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "Flashpoint Installer";
            this.Load += new System.EventHandler(this.Main_Load);
            ((System.ComponentModel.ISupportInitialize)(this.Logo)).EndInit();
            this.DestinationPathBox.ResumeLayout(false);
            this.DestinationPathBox.PerformLayout();
            this.DescriptionBox.ResumeLayout(false);
            this.ShortcutTable.ResumeLayout(false);
            this.ShortcutTable.PerformLayout();
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private System.Windows.Forms.PictureBox Logo;
        private System.Windows.Forms.Label About;
        private System.Windows.Forms.GroupBox DestinationPathBox;
        public System.Windows.Forms.TextBox DestinationPath;
        private System.Windows.Forms.Button DestinationPathBrowse;
        private System.Windows.Forms.LinkLabel Link;
        private System.Windows.Forms.Label Message;
        public RikTheVeggie.TriStateTreeView ComponentList;
        private System.Windows.Forms.TableLayoutPanel ShortcutTable;
        private System.Windows.Forms.Label ShortcutLabel;
        public System.Windows.Forms.CheckBox ShortcutStartMenu;
        public System.Windows.Forms.CheckBox ShortcutDesktop;
        private System.Windows.Forms.GroupBox DescriptionBox;
        private System.Windows.Forms.Label Description;
        private System.Windows.Forms.Label Message2;
        public System.Windows.Forms.Button InstallButton;
    }
}

