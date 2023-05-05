namespace FlashpointManager
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
            this.TabControl = new System.Windows.Forms.TabControl();
            this.ManageTab = new System.Windows.Forms.TabPage();
            this.Message = new System.Windows.Forms.Label();
            this.Message2 = new System.Windows.Forms.Label();
            this.ChangeButton = new System.Windows.Forms.Button();
            this.DescriptionBox = new System.Windows.Forms.GroupBox();
            this.Description = new System.Windows.Forms.Label();
            this.ComponentList = new RikTheVeggie.TriStateTreeView();
            this.UpdateTab = new System.Windows.Forms.TabPage();
            this.UpdateButton = new System.Windows.Forms.Button();
            this.UpdateList = new System.Windows.Forms.ListView();
            this.ComponentTitle = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.ComponentDescription = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.ComponentDate = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.ComponentSize = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.SettingsTab = new System.Windows.Forms.TabPage();
            this.SaveButton = new System.Windows.Forms.Button();
            this.BrowseButton = new System.Windows.Forms.Button();
            this.LocationBox = new System.Windows.Forms.TextBox();
            this.LocationLabel = new System.Windows.Forms.Label();
            this.CheckFilesButton = new System.Windows.Forms.Button();
            this.StableRepo = new System.Windows.Forms.RadioButton();
            this.RepositoryBox = new System.Windows.Forms.TextBox();
            this.DevRepo = new System.Windows.Forms.RadioButton();
            this.RepositoryLabel = new System.Windows.Forms.Label();
            this.CustomRepo = new System.Windows.Forms.RadioButton();
            this.UninstallButton = new System.Windows.Forms.Button();
            this.OfflineIndicator = new System.Windows.Forms.Label();
            this.TabControl.SuspendLayout();
            this.ManageTab.SuspendLayout();
            this.DescriptionBox.SuspendLayout();
            this.UpdateTab.SuspendLayout();
            this.SettingsTab.SuspendLayout();
            this.SuspendLayout();
            // 
            // TabControl
            // 
            this.TabControl.Controls.Add(this.ManageTab);
            this.TabControl.Controls.Add(this.UpdateTab);
            this.TabControl.Controls.Add(this.SettingsTab);
            this.TabControl.ItemSize = new System.Drawing.Size(150, 20);
            this.TabControl.Location = new System.Drawing.Point(13, 10);
            this.TabControl.Name = "TabControl";
            this.TabControl.SelectedIndex = 0;
            this.TabControl.Size = new System.Drawing.Size(600, 240);
            this.TabControl.SizeMode = System.Windows.Forms.TabSizeMode.Fixed;
            this.TabControl.TabIndex = 0;
            // 
            // ManageTab
            // 
            this.ManageTab.Controls.Add(this.Message);
            this.ManageTab.Controls.Add(this.Message2);
            this.ManageTab.Controls.Add(this.ChangeButton);
            this.ManageTab.Controls.Add(this.DescriptionBox);
            this.ManageTab.Controls.Add(this.ComponentList);
            this.ManageTab.Location = new System.Drawing.Point(4, 24);
            this.ManageTab.Name = "ManageTab";
            this.ManageTab.Padding = new System.Windows.Forms.Padding(3);
            this.ManageTab.Size = new System.Drawing.Size(592, 212);
            this.ManageTab.TabIndex = 1;
            this.ManageTab.Text = "Add/Remove Components";
            this.ManageTab.UseVisualStyleBackColor = true;
            // 
            // Message
            // 
            this.Message.Location = new System.Drawing.Point(410, 12);
            this.Message.Name = "Message";
            this.Message.Size = new System.Drawing.Size(171, 30);
            this.Message.TabIndex = 1;
            this.Message.Text = "Choose components to add or remove from your Flashpoint copy.";
            // 
            // Message2
            // 
            this.Message2.Location = new System.Drawing.Point(410, 44);
            this.Message2.Name = "Message2";
            this.Message2.Size = new System.Drawing.Size(171, 30);
            this.Message2.TabIndex = 2;
            this.Message2.Text = "Click on a component or category to learn more about it.";
            // 
            // ChangeButton
            // 
            this.ChangeButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.ChangeButton.Location = new System.Drawing.Point(8, 176);
            this.ChangeButton.Name = "ChangeButton";
            this.ChangeButton.Size = new System.Drawing.Size(574, 26);
            this.ChangeButton.TabIndex = 4;
            this.ChangeButton.Text = "&Apply changes";
            this.ChangeButton.UseVisualStyleBackColor = true;
            this.ChangeButton.Click += new System.EventHandler(this.ChangeButton_Click);
            // 
            // DescriptionBox
            // 
            this.DescriptionBox.Controls.Add(this.Description);
            this.DescriptionBox.Location = new System.Drawing.Point(410, 80);
            this.DescriptionBox.Name = "DescriptionBox";
            this.DescriptionBox.Padding = new System.Windows.Forms.Padding(10, 8, 10, 10);
            this.DescriptionBox.Size = new System.Drawing.Size(171, 85);
            this.DescriptionBox.TabIndex = 3;
            this.DescriptionBox.TabStop = false;
            this.DescriptionBox.Text = "Component Description";
            this.DescriptionBox.Visible = false;
            // 
            // Description
            // 
            this.Description.Dock = System.Windows.Forms.DockStyle.Fill;
            this.Description.Location = new System.Drawing.Point(10, 21);
            this.Description.Name = "Description";
            this.Description.Size = new System.Drawing.Size(151, 54);
            this.Description.TabIndex = 0;
            // 
            // ComponentList
            // 
            this.ComponentList.Indent = 20;
            this.ComponentList.Location = new System.Drawing.Point(9, 12);
            this.ComponentList.Name = "ComponentList";
            this.ComponentList.Size = new System.Drawing.Size(391, 152);
            this.ComponentList.TabIndex = 0;
            this.ComponentList.TriStateStyleProperty = RikTheVeggie.TriStateTreeView.TriStateStyles.Installer;
            this.ComponentList.AfterCheck += new System.Windows.Forms.TreeViewEventHandler(this.ComponentList_AfterCheck);
            this.ComponentList.BeforeSelect += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeSelect);
            // 
            // UpdateTab
            // 
            this.UpdateTab.Controls.Add(this.UpdateButton);
            this.UpdateTab.Controls.Add(this.UpdateList);
            this.UpdateTab.Location = new System.Drawing.Point(4, 24);
            this.UpdateTab.Name = "UpdateTab";
            this.UpdateTab.Padding = new System.Windows.Forms.Padding(3);
            this.UpdateTab.Size = new System.Drawing.Size(592, 212);
            this.UpdateTab.TabIndex = 3;
            this.UpdateTab.Text = "Update Components";
            this.UpdateTab.UseVisualStyleBackColor = true;
            // 
            // UpdateButton
            // 
            this.UpdateButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.UpdateButton.Enabled = false;
            this.UpdateButton.Location = new System.Drawing.Point(8, 176);
            this.UpdateButton.Name = "UpdateButton";
            this.UpdateButton.Size = new System.Drawing.Size(574, 26);
            this.UpdateButton.TabIndex = 1;
            this.UpdateButton.Text = "&Install updates";
            this.UpdateButton.UseVisualStyleBackColor = true;
            this.UpdateButton.Click += new System.EventHandler(this.UpdateButton_Click);
            // 
            // UpdateList
            // 
            this.UpdateList.Columns.AddRange(new System.Windows.Forms.ColumnHeader[] {
            this.ComponentTitle,
            this.ComponentDescription,
            this.ComponentDate,
            this.ComponentSize});
            this.UpdateList.FullRowSelect = true;
            this.UpdateList.GridLines = true;
            this.UpdateList.HeaderStyle = System.Windows.Forms.ColumnHeaderStyle.Nonclickable;
            this.UpdateList.HideSelection = false;
            this.UpdateList.Location = new System.Drawing.Point(9, 12);
            this.UpdateList.MultiSelect = false;
            this.UpdateList.Name = "UpdateList";
            this.UpdateList.Size = new System.Drawing.Size(572, 152);
            this.UpdateList.Sorting = System.Windows.Forms.SortOrder.Ascending;
            this.UpdateList.TabIndex = 0;
            this.UpdateList.UseCompatibleStateImageBehavior = false;
            this.UpdateList.View = System.Windows.Forms.View.Details;
            // 
            // ComponentTitle
            // 
            this.ComponentTitle.Text = "Title";
            this.ComponentTitle.Width = 120;
            // 
            // ComponentDescription
            // 
            this.ComponentDescription.Text = "Description";
            this.ComponentDescription.Width = 267;
            // 
            // ComponentDate
            // 
            this.ComponentDate.Text = "Date";
            this.ComponentDate.Width = 80;
            // 
            // ComponentSize
            // 
            this.ComponentSize.Text = "Size";
            this.ComponentSize.Width = 80;
            // 
            // SettingsTab
            // 
            this.SettingsTab.Controls.Add(this.SaveButton);
            this.SettingsTab.Controls.Add(this.BrowseButton);
            this.SettingsTab.Controls.Add(this.LocationBox);
            this.SettingsTab.Controls.Add(this.LocationLabel);
            this.SettingsTab.Controls.Add(this.CheckFilesButton);
            this.SettingsTab.Controls.Add(this.StableRepo);
            this.SettingsTab.Controls.Add(this.RepositoryBox);
            this.SettingsTab.Controls.Add(this.DevRepo);
            this.SettingsTab.Controls.Add(this.RepositoryLabel);
            this.SettingsTab.Controls.Add(this.CustomRepo);
            this.SettingsTab.Controls.Add(this.UninstallButton);
            this.SettingsTab.Location = new System.Drawing.Point(4, 24);
            this.SettingsTab.Name = "SettingsTab";
            this.SettingsTab.Padding = new System.Windows.Forms.Padding(3);
            this.SettingsTab.Size = new System.Drawing.Size(592, 212);
            this.SettingsTab.TabIndex = 2;
            this.SettingsTab.Text = "Settings";
            this.SettingsTab.UseVisualStyleBackColor = true;
            // 
            // SaveButton
            // 
            this.SaveButton.Location = new System.Drawing.Point(49, 176);
            this.SaveButton.Name = "SaveButton";
            this.SaveButton.Size = new System.Drawing.Size(160, 26);
            this.SaveButton.TabIndex = 8;
            this.SaveButton.Text = "Save and &restart";
            this.SaveButton.UseVisualStyleBackColor = true;
            this.SaveButton.Click += new System.EventHandler(this.SaveButton_Click);
            // 
            // BrowseButton
            // 
            this.BrowseButton.Location = new System.Drawing.Point(466, 43);
            this.BrowseButton.Name = "BrowseButton";
            this.BrowseButton.Size = new System.Drawing.Size(75, 22);
            this.BrowseButton.TabIndex = 2;
            this.BrowseButton.Text = "&Browse";
            this.BrowseButton.UseVisualStyleBackColor = true;
            this.BrowseButton.Click += new System.EventHandler(this.BrowseButton_Click);
            // 
            // LocationBox
            // 
            this.LocationBox.Location = new System.Drawing.Point(160, 44);
            this.LocationBox.Name = "LocationBox";
            this.LocationBox.Size = new System.Drawing.Size(300, 20);
            this.LocationBox.TabIndex = 1;
            // 
            // LocationLabel
            // 
            this.LocationLabel.AutoSize = true;
            this.LocationLabel.Location = new System.Drawing.Point(52, 47);
            this.LocationLabel.Name = "LocationLabel";
            this.LocationLabel.Size = new System.Drawing.Size(102, 13);
            this.LocationLabel.TabIndex = 0;
            this.LocationLabel.Text = "Flashpoint Location:";
            // 
            // CheckFilesButton
            // 
            this.CheckFilesButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.CheckFilesButton.Location = new System.Drawing.Point(215, 176);
            this.CheckFilesButton.Name = "CheckFilesButton";
            this.CheckFilesButton.Size = new System.Drawing.Size(160, 26);
            this.CheckFilesButton.TabIndex = 9;
            this.CheckFilesButton.Text = "Check for &missing files";
            this.CheckFilesButton.UseVisualStyleBackColor = true;
            this.CheckFilesButton.Click += new System.EventHandler(this.CheckFilesButton_Click);
            // 
            // StableRepo
            // 
            this.StableRepo.AutoSize = true;
            this.StableRepo.Location = new System.Drawing.Point(203, 122);
            this.StableRepo.Name = "StableRepo";
            this.StableRepo.Size = new System.Drawing.Size(55, 17);
            this.StableRepo.TabIndex = 5;
            this.StableRepo.TabStop = true;
            this.StableRepo.Text = "&Stable";
            this.StableRepo.UseVisualStyleBackColor = true;
            this.StableRepo.CheckedChanged += new System.EventHandler(this.ComponentRepo_CheckedChanged);
            // 
            // RepositoryBox
            // 
            this.RepositoryBox.Location = new System.Drawing.Point(160, 96);
            this.RepositoryBox.Name = "RepositoryBox";
            this.RepositoryBox.Size = new System.Drawing.Size(300, 20);
            this.RepositoryBox.TabIndex = 4;
            // 
            // DevRepo
            // 
            this.DevRepo.AutoSize = true;
            this.DevRepo.Location = new System.Drawing.Point(264, 122);
            this.DevRepo.Name = "DevRepo";
            this.DevRepo.Size = new System.Drawing.Size(88, 17);
            this.DevRepo.TabIndex = 6;
            this.DevRepo.TabStop = true;
            this.DevRepo.Text = "&Development";
            this.DevRepo.UseVisualStyleBackColor = true;
            this.DevRepo.CheckedChanged += new System.EventHandler(this.ComponentRepo_CheckedChanged);
            // 
            // RepositoryLabel
            // 
            this.RepositoryLabel.AutoSize = true;
            this.RepositoryLabel.Location = new System.Drawing.Point(52, 99);
            this.RepositoryLabel.Name = "RepositoryLabel";
            this.RepositoryLabel.Size = new System.Drawing.Size(101, 13);
            this.RepositoryLabel.TabIndex = 3;
            this.RepositoryLabel.Text = "Component Source:";
            // 
            // CustomRepo
            // 
            this.CustomRepo.AutoSize = true;
            this.CustomRepo.Location = new System.Drawing.Point(358, 122);
            this.CustomRepo.Name = "CustomRepo";
            this.CustomRepo.Size = new System.Drawing.Size(60, 17);
            this.CustomRepo.TabIndex = 7;
            this.CustomRepo.TabStop = true;
            this.CustomRepo.Text = "&Custom";
            this.CustomRepo.UseVisualStyleBackColor = true;
            this.CustomRepo.CheckedChanged += new System.EventHandler(this.ComponentRepo_CheckedChanged);
            // 
            // UninstallButton
            // 
            this.UninstallButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.UninstallButton.Location = new System.Drawing.Point(381, 176);
            this.UninstallButton.Name = "UninstallButton";
            this.UninstallButton.Size = new System.Drawing.Size(160, 26);
            this.UninstallButton.TabIndex = 10;
            this.UninstallButton.Text = "&Uninstall Flashpoint";
            this.UninstallButton.UseVisualStyleBackColor = true;
            this.UninstallButton.Click += new System.EventHandler(this.UninstallButton_Click);
            // 
            // OfflineIndicator
            // 
            this.OfflineIndicator.AutoSize = true;
            this.OfflineIndicator.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Bold, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.OfflineIndicator.Location = new System.Drawing.Point(530, 16);
            this.OfflineIndicator.Name = "OfflineIndicator";
            this.OfflineIndicator.Size = new System.Drawing.Size(79, 13);
            this.OfflineIndicator.TabIndex = 1;
            this.OfflineIndicator.Text = "Offline Mode";
            this.OfflineIndicator.Visible = false;
            // 
            // Main
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(624, 261);
            this.Controls.Add(this.OfflineIndicator);
            this.Controls.Add(this.TabControl);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.Icon = ((System.Drawing.Icon)(resources.GetObject("$this.Icon")));
            this.MaximizeBox = false;
            this.Name = "Main";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "Flashpoint Manager";
            this.FormClosing += new System.Windows.Forms.FormClosingEventHandler(this.Main_FormClosing);
            this.Load += new System.EventHandler(this.Main_Load);
            this.TabControl.ResumeLayout(false);
            this.ManageTab.ResumeLayout(false);
            this.DescriptionBox.ResumeLayout(false);
            this.UpdateTab.ResumeLayout(false);
            this.SettingsTab.ResumeLayout(false);
            this.SettingsTab.PerformLayout();
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion
        private System.Windows.Forms.TabControl TabControl;
        private System.Windows.Forms.TabPage ManageTab;
        private System.Windows.Forms.TabPage SettingsTab;
        public RikTheVeggie.TriStateTreeView ComponentList;
        private System.Windows.Forms.GroupBox DescriptionBox;
        private System.Windows.Forms.Label Description;
        private System.Windows.Forms.Label Message;
        public System.Windows.Forms.Button ChangeButton;
        public System.Windows.Forms.Button UninstallButton;
        public System.Windows.Forms.Label Message2;
        private System.Windows.Forms.TabPage UpdateTab;
        private System.Windows.Forms.ColumnHeader ComponentTitle;
        private System.Windows.Forms.ColumnHeader ComponentDescription;
        private System.Windows.Forms.ColumnHeader ComponentDate;
        private System.Windows.Forms.ColumnHeader ComponentSize;
        public System.Windows.Forms.Button UpdateButton;
        public System.Windows.Forms.ListView UpdateList;
        private System.Windows.Forms.Label RepositoryLabel;
        private System.Windows.Forms.Button BrowseButton;
        private System.Windows.Forms.Label LocationLabel;
        public System.Windows.Forms.Button CheckFilesButton;
        private System.Windows.Forms.Button SaveButton;
        public System.Windows.Forms.TextBox RepositoryBox;
        public System.Windows.Forms.TextBox LocationBox;
        public System.Windows.Forms.RadioButton StableRepo;
        public System.Windows.Forms.RadioButton DevRepo;
        public System.Windows.Forms.RadioButton CustomRepo;
        private System.Windows.Forms.Label OfflineIndicator;
    }
}

