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
            this.FolderText = new System.Windows.Forms.TextBox();
            this.FolderButton = new System.Windows.Forms.Button();
            this.ShortcutsText = new System.Windows.Forms.Label();
            this.ShortcutsDesktop = new System.Windows.Forms.CheckBox();
            this.Shortcuts = new System.Windows.Forms.TableLayoutPanel();
            this.ShortcutsStart = new System.Windows.Forms.CheckBox();
            this.Install = new System.Windows.Forms.Button();
            this.About = new System.Windows.Forms.Label();
            ((System.ComponentModel.ISupportInitialize)(this.Logo)).BeginInit();
            this.Folder.SuspendLayout();
            this.Shortcuts.SuspendLayout();
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
            this.Folder.Controls.Add(this.FolderText);
            this.Folder.Controls.Add(this.FolderButton);
            this.Folder.Location = new System.Drawing.Point(12, 230);
            this.Folder.Name = "Folder";
            this.Folder.Size = new System.Drawing.Size(600, 51);
            this.Folder.TabIndex = 2;
            this.Folder.TabStop = false;
            this.Folder.Text = "Containing folder:";
            // 
            // FolderText
            // 
            this.FolderText.Location = new System.Drawing.Point(6, 19);
            this.FolderText.Name = "FolderText";
            this.FolderText.Size = new System.Drawing.Size(507, 23);
            this.FolderText.TabIndex = 2;
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
            // ShortcutsText
            // 
            this.ShortcutsText.Dock = System.Windows.Forms.DockStyle.Fill;
            this.ShortcutsText.Location = new System.Drawing.Point(3, 0);
            this.ShortcutsText.Name = "ShortcutsText";
            this.ShortcutsText.Padding = new System.Windows.Forms.Padding(0, 4, 0, 0);
            this.ShortcutsText.Size = new System.Drawing.Size(127, 25);
            this.ShortcutsText.TabIndex = 0;
            this.ShortcutsText.Text = "Create shortcuts in:";
            // 
            // ShortcutsDesktop
            // 
            this.ShortcutsDesktop.CheckAlign = System.Drawing.ContentAlignment.TopLeft;
            this.ShortcutsDesktop.Dock = System.Windows.Forms.DockStyle.Fill;
            this.ShortcutsDesktop.Location = new System.Drawing.Point(136, 3);
            this.ShortcutsDesktop.Name = "ShortcutsDesktop";
            this.ShortcutsDesktop.Size = new System.Drawing.Size(78, 19);
            this.ShortcutsDesktop.TabIndex = 1;
            this.ShortcutsDesktop.Text = "Desktop";
            this.ShortcutsDesktop.TextAlign = System.Drawing.ContentAlignment.TopLeft;
            this.ShortcutsDesktop.UseVisualStyleBackColor = true;
            // 
            // Shortcuts
            // 
            this.Shortcuts.ColumnCount = 3;
            this.Shortcuts.ColumnStyles.Add(new System.Windows.Forms.ColumnStyle(System.Windows.Forms.SizeType.Absolute, 133F));
            this.Shortcuts.ColumnStyles.Add(new System.Windows.Forms.ColumnStyle(System.Windows.Forms.SizeType.Absolute, 84F));
            this.Shortcuts.ColumnStyles.Add(new System.Windows.Forms.ColumnStyle(System.Windows.Forms.SizeType.Absolute, 26F));
            this.Shortcuts.Controls.Add(this.ShortcutsText, 0, 0);
            this.Shortcuts.Controls.Add(this.ShortcutsDesktop, 1, 0);
            this.Shortcuts.Controls.Add(this.ShortcutsStart, 2, 0);
            this.Shortcuts.Location = new System.Drawing.Point(12, 289);
            this.Shortcuts.Name = "Shortcuts";
            this.Shortcuts.RowCount = 1;
            this.Shortcuts.RowStyles.Add(new System.Windows.Forms.RowStyle(System.Windows.Forms.SizeType.Percent, 100F));
            this.Shortcuts.Size = new System.Drawing.Size(318, 25);
            this.Shortcuts.TabIndex = 4;
            // 
            // ShortcutsStart
            // 
            this.ShortcutsStart.CheckAlign = System.Drawing.ContentAlignment.TopLeft;
            this.ShortcutsStart.Dock = System.Windows.Forms.DockStyle.Fill;
            this.ShortcutsStart.Location = new System.Drawing.Point(220, 3);
            this.ShortcutsStart.Name = "ShortcutsStart";
            this.ShortcutsStart.Size = new System.Drawing.Size(95, 19);
            this.ShortcutsStart.TabIndex = 2;
            this.ShortcutsStart.Text = "Start menu";
            this.ShortcutsStart.TextAlign = System.Drawing.ContentAlignment.TopLeft;
            this.ShortcutsStart.UseVisualStyleBackColor = true;
            // 
            // Install
            // 
            this.Install.Location = new System.Drawing.Point(336, 287);
            this.Install.Name = "Install";
            this.Install.Size = new System.Drawing.Size(276, 27);
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
            this.Controls.Add(this.Shortcuts);
            this.Controls.Add(this.Install);
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
            this.Shortcuts.ResumeLayout(false);
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private PictureBox Logo;
        private LinkLabel Link;
        private GroupBox Folder;
        private Button FolderButton;
        private Label ShortcutsText;
        private TableLayoutPanel Shortcuts;
        private Button Install;
        private Label About;
        public TextBox FolderText;
        public CheckBox ShortcutsDesktop;
        public CheckBox ShortcutsStart;
    }
}