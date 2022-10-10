namespace FlashpointInstaller
{
    partial class Main
    {
        /// <summary>
        ///  Required designer variable.
        /// </summary>
        private System.ComponentModel.IContainer components = null;

        /// <summary>
        ///  Clean up any resources being used.
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
        ///  Required method for Designer support - do not modify
        ///  the contents of this method with the code editor.
        /// </summary>
        private void InitializeComponent()
        {
            System.ComponentModel.ComponentResourceManager resources = new System.ComponentModel.ComponentResourceManager(typeof(Main));
            this.Logo = new System.Windows.Forms.PictureBox();
            this.Link = new System.Windows.Forms.LinkLabel();
            this.Folder = new System.Windows.Forms.GroupBox();
            this.FolderTextBox = new System.Windows.Forms.TextBox();
            this.FolderButton = new System.Windows.Forms.Button();
            this.Shortcut = new System.Windows.Forms.CheckBox();
            this.Install = new System.Windows.Forms.Button();
            this.About = new System.Windows.Forms.Label();
            ((System.ComponentModel.ISupportInitialize)(this.Logo)).BeginInit();
            this.Folder.SuspendLayout();
            this.SuspendLayout();
            // 
            // Logo
            // 
            this.Logo.Image = global::FlashpointInstaller.Properties.Resources.FlashpointLogo;
            this.Logo.Location = new System.Drawing.Point(12, 12);
            this.Logo.Name = "Logo";
            this.Logo.Size = new System.Drawing.Size(600, 177);
            this.Logo.SizeMode = System.Windows.Forms.PictureBoxSizeMode.CenterImage;
            this.Logo.TabIndex = 0;
            this.Logo.TabStop = false;
            // 
            // Link
            // 
            this.Link.AutoSize = true;
            this.Link.LinkColor = System.Drawing.Color.FromArgb(((int)(((byte)(4)))), ((int)(((byte)(94)))), ((int)(((byte)(221)))));
            this.Link.Location = new System.Drawing.Point(415, 192);
            this.Link.Name = "Link";
            this.Link.Size = new System.Drawing.Size(197, 15);
            this.Link.TabIndex = 1;
            this.Link.TabStop = true;
            this.Link.Text = "https://bluemaxima.org/flashpoint/";
            this.Link.LinkClicked += new System.Windows.Forms.LinkLabelLinkClickedEventHandler(this.Link_LinkClicked);
            // 
            // Folder
            // 
            this.Folder.Controls.Add(this.FolderTextBox);
            this.Folder.Controls.Add(this.FolderButton);
            this.Folder.Location = new System.Drawing.Point(12, 230);
            this.Folder.Name = "Folder";
            this.Folder.Size = new System.Drawing.Size(600, 51);
            this.Folder.TabIndex = 2;
            this.Folder.TabStop = false;
            this.Folder.Text = "Containing folder:";
            // 
            // FolderTextBox
            // 
            this.FolderTextBox.Location = new System.Drawing.Point(6, 19);
            this.FolderTextBox.Name = "FolderTextBox";
            this.FolderTextBox.Size = new System.Drawing.Size(507, 23);
            this.FolderTextBox.TabIndex = 2;
            // 
            // FolderButton
            // 
            this.FolderButton.Location = new System.Drawing.Point(519, 18);
            this.FolderButton.Name = "FolderButton";
            this.FolderButton.Size = new System.Drawing.Size(75, 23);
            this.FolderButton.TabIndex = 1;
            this.FolderButton.Text = "Browse";
            this.FolderButton.UseVisualStyleBackColor = true;
            this.FolderButton.Click += new System.EventHandler(this.FolderButton_Click);
            // 
            // Shortcut
            // 
            this.Shortcut.AutoSize = true;
            this.Shortcut.CheckAlign = System.Drawing.ContentAlignment.TopLeft;
            this.Shortcut.Checked = true;
            this.Shortcut.CheckState = System.Windows.Forms.CheckState.Checked;
            this.Shortcut.Location = new System.Drawing.Point(13, 292);
            this.Shortcut.Name = "Shortcut";
            this.Shortcut.Size = new System.Drawing.Size(169, 19);
            this.Shortcut.TabIndex = 1;
            this.Shortcut.Text = "Create shortcut on desktop";
            this.Shortcut.TextAlign = System.Drawing.ContentAlignment.TopLeft;
            this.Shortcut.UseVisualStyleBackColor = true;
            // 
            // Install
            // 
            this.Install.Location = new System.Drawing.Point(193, 287);
            this.Install.Name = "Install";
            this.Install.Size = new System.Drawing.Size(419, 27);
            this.Install.TabIndex = 5;
            this.Install.Text = "Install Flashpoint Infinity";
            this.Install.UseVisualStyleBackColor = true;
            this.Install.Click += new System.EventHandler(this.Install_Click);
            // 
            // About
            // 
            this.About.AutoSize = true;
            this.About.Location = new System.Drawing.Point(12, 192);
            this.About.Name = "About";
            this.About.Size = new System.Drawing.Size(170, 15);
            this.About.TabIndex = 6;
            this.About.Text = "Flashpoint Infinity Installer v1.0";
            // 
            // Main
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(7F, 15F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(624, 331);
            this.Controls.Add(this.Logo);
            this.Controls.Add(this.About);
            this.Controls.Add(this.Link);
            this.Controls.Add(this.Folder);
            this.Controls.Add(this.Install);
            this.Controls.Add(this.Shortcut);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.Icon = ((System.Drawing.Icon)(resources.GetObject("$this.Icon")));
            this.MaximizeBox = false;
            this.Name = "Main";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "Flashpoint Infinity Installer";
            this.Load += new System.EventHandler(this.Main_Load);
            ((System.ComponentModel.ISupportInitialize)(this.Logo)).EndInit();
            this.Folder.ResumeLayout(false);
            this.Folder.PerformLayout();
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private PictureBox Logo;
        private LinkLabel Link;
        private GroupBox Folder;
        private Button FolderButton;
        private Button Install;
        private Label About;
        public TextBox FolderTextBox;
        public CheckBox Shortcut;
    }
}