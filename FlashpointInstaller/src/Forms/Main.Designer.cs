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
            this.Logo.Location = new System.Drawing.Point(32, 18);
            this.Logo.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.Logo.Name = "Logo";
            this.Logo.Size = new System.Drawing.Size(886, 272);
            this.Logo.SizeMode = System.Windows.Forms.PictureBoxSizeMode.CenterImage;
            this.Logo.TabIndex = 0;
            this.Logo.TabStop = false;
            // 
            // About
            // 
            this.About.AutoSize = true;
            this.About.Location = new System.Drawing.Point(27, 295);
            this.About.Margin = new System.Windows.Forms.Padding(4, 0, 4, 0);
            this.About.Name = "About";
            this.About.Size = new System.Drawing.Size(143, 20);
            this.About.TabIndex = 0;
            this.About.Text = "Flashpoint Installer";
            // 
            // DestinationPathBox
            // 
            this.DestinationPathBox.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.DestinationPathBox.Controls.Add(this.DestinationPath);
            this.DestinationPathBox.Controls.Add(this.DestinationPathBrowse);
            this.DestinationPathBox.Location = new System.Drawing.Point(18, 591);
            this.DestinationPathBox.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.DestinationPathBox.Name = "DestinationPathBox";
            this.DestinationPathBox.Padding = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.DestinationPathBox.Size = new System.Drawing.Size(900, 75);
            this.DestinationPathBox.TabIndex = 6;
            this.DestinationPathBox.TabStop = false;
            this.DestinationPathBox.Text = "Destination Folder";
            // 
            // DestinationPath
            // 
            this.DestinationPath.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Regular, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.DestinationPath.Location = new System.Drawing.Point(14, 28);
            this.DestinationPath.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.DestinationPath.Name = "DestinationPath";
            this.DestinationPath.Size = new System.Drawing.Size(752, 26);
            this.DestinationPath.TabIndex = 0;
            this.DestinationPath.KeyPress += new System.Windows.Forms.KeyPressEventHandler(this.DestinationPath_KeyPress);
            // 
            // DestinationPathBrowse
            // 
            this.DestinationPathBrowse.Location = new System.Drawing.Point(778, 26);
            this.DestinationPathBrowse.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.DestinationPathBrowse.Name = "DestinationPathBrowse";
            this.DestinationPathBrowse.Size = new System.Drawing.Size(110, 34);
            this.DestinationPathBrowse.TabIndex = 1;
            this.DestinationPathBrowse.Text = "&Browse";
            this.DestinationPathBrowse.UseVisualStyleBackColor = true;
            this.DestinationPathBrowse.Click += new System.EventHandler(this.DestinationPathBrowse_Click);
            // 
            // InstallButton
            // 
            this.InstallButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.InstallButton.Location = new System.Drawing.Point(450, 682);
            this.InstallButton.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.InstallButton.Name = "InstallButton";
            this.InstallButton.Size = new System.Drawing.Size(470, 40);
            this.InstallButton.TabIndex = 8;
            this.InstallButton.Text = "&Install Flashpoint";
            this.InstallButton.UseVisualStyleBackColor = true;
            this.InstallButton.Click += new System.EventHandler(this.InstallButton_Click);
            // 
            // Link
            // 
            this.Link.AutoSize = true;
            this.Link.LinkColor = System.Drawing.Color.FromArgb(((int)(((byte)(4)))), ((int)(((byte)(94)))), ((int)(((byte)(221)))));
            this.Link.Location = new System.Drawing.Point(713, 295);
            this.Link.Margin = new System.Windows.Forms.Padding(4, 0, 4, 0);
            this.Link.Name = "Link";
            this.Link.Size = new System.Drawing.Size(207, 20);
            this.Link.TabIndex = 1;
            this.Link.TabStop = true;
            this.Link.Text = "https://flashpointarchive.org/";
            this.Link.TextAlign = System.Drawing.ContentAlignment.TopRight;
            this.Link.LinkClicked += new System.Windows.Forms.LinkLabelLinkClickedEventHandler(this.Link_LinkClicked);
            // 
            // Message2
            // 
            this.Message2.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.Message2.Location = new System.Drawing.Point(622, 395);
            this.Message2.Margin = new System.Windows.Forms.Padding(4, 0, 4, 0);
            this.Message2.Name = "Message2";
            this.Message2.Size = new System.Drawing.Size(296, 46);
            this.Message2.TabIndex = 4;
            this.Message2.Text = "Click on a component or category to learn more about it.";
            // 
            // DescriptionBox
            // 
            this.DescriptionBox.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.DescriptionBox.Controls.Add(this.Description);
            this.DescriptionBox.Location = new System.Drawing.Point(622, 469);
            this.DescriptionBox.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.DescriptionBox.Name = "DescriptionBox";
            this.DescriptionBox.Padding = new System.Windows.Forms.Padding(15, 12, 15, 15);
            this.DescriptionBox.Size = new System.Drawing.Size(296, 111);
            this.DescriptionBox.TabIndex = 5;
            this.DescriptionBox.TabStop = false;
            this.DescriptionBox.Text = "Component Description";
            this.DescriptionBox.Visible = false;
            // 
            // Description
            // 
            this.Description.Location = new System.Drawing.Point(15, 32);
            this.Description.Margin = new System.Windows.Forms.Padding(4, 0, 4, 0);
            this.Description.Name = "Description";
            this.Description.Size = new System.Drawing.Size(266, 63);
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
            this.ShortcutTable.Location = new System.Drawing.Point(18, 680);
            this.ShortcutTable.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.ShortcutTable.Name = "ShortcutTable";
            this.ShortcutTable.RowCount = 1;
            this.ShortcutTable.RowStyles.Add(new System.Windows.Forms.RowStyle(System.Windows.Forms.SizeType.Percent, 100F));
            this.ShortcutTable.Size = new System.Drawing.Size(411, 40);
            this.ShortcutTable.TabIndex = 7;
            // 
            // ShortcutStartMenu
            // 
            this.ShortcutStartMenu.AutoSize = true;
            this.ShortcutStartMenu.Checked = true;
            this.ShortcutStartMenu.CheckState = System.Windows.Forms.CheckState.Checked;
            this.ShortcutStartMenu.Location = new System.Drawing.Point(293, 5);
            this.ShortcutStartMenu.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.ShortcutStartMenu.Name = "ShortcutStartMenu";
            this.ShortcutStartMenu.Padding = new System.Windows.Forms.Padding(0, 5, 0, 0);
            this.ShortcutStartMenu.Size = new System.Drawing.Size(114, 29);
            this.ShortcutStartMenu.TabIndex = 2;
            this.ShortcutStartMenu.Text = "&Start Menu";
            this.ShortcutStartMenu.UseVisualStyleBackColor = true;
            // 
            // ShortcutDesktop
            // 
            this.ShortcutDesktop.AutoSize = true;
            this.ShortcutDesktop.Checked = true;
            this.ShortcutDesktop.CheckState = System.Windows.Forms.CheckState.Checked;
            this.ShortcutDesktop.Location = new System.Drawing.Point(190, 5);
            this.ShortcutDesktop.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.ShortcutDesktop.Name = "ShortcutDesktop";
            this.ShortcutDesktop.Padding = new System.Windows.Forms.Padding(0, 5, 0, 0);
            this.ShortcutDesktop.Size = new System.Drawing.Size(95, 29);
            this.ShortcutDesktop.TabIndex = 1;
            this.ShortcutDesktop.Text = "&Desktop";
            this.ShortcutDesktop.UseVisualStyleBackColor = true;
            // 
            // ShortcutLabel
            // 
            this.ShortcutLabel.AutoSize = true;
            this.ShortcutLabel.Dock = System.Windows.Forms.DockStyle.Fill;
            this.ShortcutLabel.Location = new System.Drawing.Point(4, 5);
            this.ShortcutLabel.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.ShortcutLabel.Name = "ShortcutLabel";
            this.ShortcutLabel.Size = new System.Drawing.Size(178, 30);
            this.ShortcutLabel.TabIndex = 0;
            this.ShortcutLabel.Text = "Create shortcuts in:";
            this.ShortcutLabel.TextAlign = System.Drawing.ContentAlignment.MiddleCenter;
            // 
            // Message
            // 
            this.Message.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.Message.Location = new System.Drawing.Point(622, 346);
            this.Message.Margin = new System.Windows.Forms.Padding(4, 0, 4, 0);
            this.Message.Name = "Message";
            this.Message.Size = new System.Drawing.Size(296, 46);
            this.Message.TabIndex = 3;
            this.Message.Text = "Choose components to include in your Flashpoint installation.";
            // 
            // ComponentList
            // 
            this.ComponentList.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.ComponentList.Indent = 20;
            this.ComponentList.Location = new System.Drawing.Point(32, 338);
            this.ComponentList.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.ComponentList.Name = "ComponentList";
            this.ComponentList.Size = new System.Drawing.Size(574, 238);
            this.ComponentList.TabIndex = 2;
            this.ComponentList.TriStateStyleProperty = RikTheVeggie.TriStateTreeView.TriStateStyles.Installer;
            this.ComponentList.BeforeCheck += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeCheck);
            this.ComponentList.AfterCheck += new System.Windows.Forms.TreeViewEventHandler(this.ComponentList_AfterCheck);
            this.ComponentList.BeforeSelect += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeSelect);
            // 
            // Main
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(9F, 20F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(936, 740);
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
            this.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
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

