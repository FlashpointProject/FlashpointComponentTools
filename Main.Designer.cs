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
            System.ComponentModel.ComponentResourceManager resources = new System.ComponentModel.ComponentResourceManager(typeof(Main));
            this.Logo = new System.Windows.Forms.PictureBox();
            this.About = new System.Windows.Forms.Label();
            this.Folder = new System.Windows.Forms.GroupBox();
            this.FolderButton = new System.Windows.Forms.Button();
            this.FolderTextBox = new System.Windows.Forms.TextBox();
            this.Install = new System.Windows.Forms.Button();
            this.Shortcut = new System.Windows.Forms.CheckBox();
            this.Link = new System.Windows.Forms.LinkLabel();
            ((System.ComponentModel.ISupportInitialize)(this.Logo)).BeginInit();
            this.Folder.SuspendLayout();
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
            this.About.Size = new System.Drawing.Size(151, 13);
            this.About.TabIndex = 1;
            this.About.Text = "Flashpoint Infinity Installer v1.0";
            // 
            // Folder
            // 
            this.Folder.Controls.Add(this.FolderTextBox);
            this.Folder.Controls.Add(this.FolderButton);
            this.Folder.Location = new System.Drawing.Point(12, 228);
            this.Folder.Name = "Folder";
            this.Folder.Size = new System.Drawing.Size(600, 49);
            this.Folder.TabIndex = 2;
            this.Folder.TabStop = false;
            this.Folder.Text = "Containing folder:";
            // 
            // FolderButton
            // 
            this.FolderButton.Location = new System.Drawing.Point(519, 16);
            this.FolderButton.Name = "FolderButton";
            this.FolderButton.Size = new System.Drawing.Size(73, 22);
            this.FolderButton.TabIndex = 0;
            this.FolderButton.Text = "Browse";
            this.FolderButton.UseVisualStyleBackColor = true;
            this.FolderButton.Click += new System.EventHandler(this.FolderButton_Click);
            // 
            // FolderTextBox
            // 
            this.FolderTextBox.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Regular, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.FolderTextBox.Location = new System.Drawing.Point(9, 17);
            this.FolderTextBox.Name = "FolderTextBox";
            this.FolderTextBox.Size = new System.Drawing.Size(503, 20);
            this.FolderTextBox.TabIndex = 1;
            // 
            // Install
            // 
            this.Install.Location = new System.Drawing.Point(192, 283);
            this.Install.Name = "Install";
            this.Install.Size = new System.Drawing.Size(420, 26);
            this.Install.TabIndex = 2;
            this.Install.Text = "Install Flashpoint Infinity";
            this.Install.UseVisualStyleBackColor = true;
            this.Install.Click += new System.EventHandler(this.Install_Click);
            // 
            // Shortcut
            // 
            this.Shortcut.AutoSize = true;
            this.Shortcut.Checked = true;
            this.Shortcut.CheckState = System.Windows.Forms.CheckState.Checked;
            this.Shortcut.Location = new System.Drawing.Point(15, 288);
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
            // Main
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(624, 321);
            this.Controls.Add(this.Link);
            this.Controls.Add(this.Shortcut);
            this.Controls.Add(this.Install);
            this.Controls.Add(this.Folder);
            this.Controls.Add(this.About);
            this.Controls.Add(this.Logo);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.Icon = ((System.Drawing.Icon)(resources.GetObject("$this.Icon")));
            this.MaximizeBox = false;
            this.MinimizeBox = false;
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

        private System.Windows.Forms.PictureBox Logo;
        private System.Windows.Forms.Label About;
        private System.Windows.Forms.GroupBox Folder;
        public System.Windows.Forms.TextBox FolderTextBox;
        private System.Windows.Forms.Button FolderButton;
        private System.Windows.Forms.Button Install;
        private System.Windows.Forms.LinkLabel Link;
        public System.Windows.Forms.CheckBox Shortcut;
    }
}

