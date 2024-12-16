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
            this.chkUncheckAll = new System.Windows.Forms.CheckBox();
            this.groupBox1 = new System.Windows.Forms.GroupBox();
            this.lblTotalUpdates = new System.Windows.Forms.Label();
            this.lblTotalUpdatesSize = new System.Windows.Forms.Label();
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
            this.groupBox1.SuspendLayout();
            this.SettingsTab.SuspendLayout();
            this.SuspendLayout();
            // 
            // TabControl
            // 
            this.TabControl.Controls.Add(this.ManageTab);
            this.TabControl.Controls.Add(this.UpdateTab);
            this.TabControl.Controls.Add(this.SettingsTab);
            this.TabControl.ItemSize = new System.Drawing.Size(150, 20);
            this.TabControl.Location = new System.Drawing.Point(20, 15);
            this.TabControl.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.TabControl.Name = "TabControl";
            this.TabControl.SelectedIndex = 0;
            this.TabControl.Size = new System.Drawing.Size(916, 435);
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
            this.ManageTab.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.ManageTab.Name = "ManageTab";
            this.ManageTab.Padding = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.ManageTab.Size = new System.Drawing.Size(908, 407);
            this.ManageTab.TabIndex = 1;
            this.ManageTab.Text = "Add/Remove Components";
            this.ManageTab.UseVisualStyleBackColor = true;
            // 
            // Message
            // 
            this.Message.Location = new System.Drawing.Point(615, 18);
            this.Message.Margin = new System.Windows.Forms.Padding(4, 0, 4, 0);
            this.Message.Name = "Message";
            this.Message.Size = new System.Drawing.Size(256, 46);
            this.Message.TabIndex = 1;
            this.Message.Text = "Choose components to add or remove from your Flashpoint copy.";
            // 
            // Message2
            // 
            this.Message2.Location = new System.Drawing.Point(615, 68);
            this.Message2.Margin = new System.Windows.Forms.Padding(4, 0, 4, 0);
            this.Message2.Name = "Message2";
            this.Message2.Size = new System.Drawing.Size(256, 46);
            this.Message2.TabIndex = 2;
            this.Message2.Text = "Click on a component or category to learn more about it.";
            // 
            // ChangeButton
            // 
            this.ChangeButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.ChangeButton.Location = new System.Drawing.Point(20, 337);
            this.ChangeButton.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.ChangeButton.Name = "ChangeButton";
            this.ChangeButton.Size = new System.Drawing.Size(861, 40);
            this.ChangeButton.TabIndex = 4;
            this.ChangeButton.Text = "&Apply changes";
            this.ChangeButton.UseVisualStyleBackColor = true;
            this.ChangeButton.Click += new System.EventHandler(this.ChangeButton_Click);
            // 
            // DescriptionBox
            // 
            this.DescriptionBox.Controls.Add(this.Description);
            this.DescriptionBox.Location = new System.Drawing.Point(615, 123);
            this.DescriptionBox.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.DescriptionBox.Name = "DescriptionBox";
            this.DescriptionBox.Padding = new System.Windows.Forms.Padding(15, 12, 15, 15);
            this.DescriptionBox.Size = new System.Drawing.Size(256, 131);
            this.DescriptionBox.TabIndex = 3;
            this.DescriptionBox.TabStop = false;
            this.DescriptionBox.Text = "Component Description";
            this.DescriptionBox.Visible = false;
            // 
            // Description
            // 
            this.Description.Dock = System.Windows.Forms.DockStyle.Fill;
            this.Description.Location = new System.Drawing.Point(15, 31);
            this.Description.Margin = new System.Windows.Forms.Padding(4, 0, 4, 0);
            this.Description.Name = "Description";
            this.Description.Size = new System.Drawing.Size(226, 85);
            this.Description.TabIndex = 0;
            // 
            // ComponentList
            // 
            this.ComponentList.Indent = 20;
            this.ComponentList.Location = new System.Drawing.Point(14, 18);
            this.ComponentList.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.ComponentList.Name = "ComponentList";
            this.ComponentList.Size = new System.Drawing.Size(584, 232);
            this.ComponentList.TabIndex = 0;
            this.ComponentList.TriStateStyleProperty = RikTheVeggie.TriStateTreeView.TriStateStyles.Installer;
            this.ComponentList.AfterCheck += new System.Windows.Forms.TreeViewEventHandler(this.ComponentList_AfterCheck);
            this.ComponentList.BeforeSelect += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeSelect);
            // 
            // UpdateTab
            // 
            this.UpdateTab.Controls.Add(this.chkUncheckAll);
            this.UpdateTab.Controls.Add(this.groupBox1);
            this.UpdateTab.Controls.Add(this.UpdateButton);
            this.UpdateTab.Controls.Add(this.UpdateList);
            this.UpdateTab.Location = new System.Drawing.Point(4, 24);
            this.UpdateTab.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.UpdateTab.Name = "UpdateTab";
            this.UpdateTab.Padding = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.UpdateTab.Size = new System.Drawing.Size(908, 407);
            this.UpdateTab.TabIndex = 3;
            this.UpdateTab.Text = "Update Components";
            this.UpdateTab.UseVisualStyleBackColor = true;
            // 
            // chkUncheckAll
            // 
            this.chkUncheckAll.AutoSize = true;
            this.chkUncheckAll.Checked = true;
            this.chkUncheckAll.CheckState = System.Windows.Forms.CheckState.Checked;
            this.chkUncheckAll.Location = new System.Drawing.Point(8, 55);
            this.chkUncheckAll.Name = "chkUncheckAll";
            this.chkUncheckAll.Size = new System.Drawing.Size(266, 24);
            this.chkUncheckAll.TabIndex = 3;
            this.chkUncheckAll.Text = "Check / Uncheck all components";
            this.chkUncheckAll.UseVisualStyleBackColor = true;
            this.chkUncheckAll.CheckedChanged += new System.EventHandler(this.chkUncheckAll_CheckedChanged);
            // 
            // groupBox1
            // 
            this.groupBox1.Controls.Add(this.lblTotalUpdates);
            this.groupBox1.Controls.Add(this.lblTotalUpdatesSize);
            this.groupBox1.Location = new System.Drawing.Point(649, 9);
            this.groupBox1.Name = "groupBox1";
            this.groupBox1.Size = new System.Drawing.Size(235, 75);
            this.groupBox1.TabIndex = 2;
            this.groupBox1.TabStop = false;
            this.groupBox1.Text = "Update Info";
            // 
            // lblTotalUpdates
            // 
            this.lblTotalUpdates.AutoSize = true;
            this.lblTotalUpdates.Location = new System.Drawing.Point(6, 50);
            this.lblTotalUpdates.Name = "lblTotalUpdates";
            this.lblTotalUpdates.Size = new System.Drawing.Size(120, 20);
            this.lblTotalUpdates.TabIndex = 3;
            this.lblTotalUpdates.Text = "lblTotalUpdates";
            // 
            // lblTotalUpdatesSize
            // 
            this.lblTotalUpdatesSize.AutoSize = true;
            this.lblTotalUpdatesSize.Location = new System.Drawing.Point(6, 22);
            this.lblTotalUpdatesSize.Name = "lblTotalUpdatesSize";
            this.lblTotalUpdatesSize.Size = new System.Drawing.Size(151, 20);
            this.lblTotalUpdatesSize.TabIndex = 0;
            this.lblTotalUpdatesSize.Text = "lblTotalUpdatesSize";
            // 
            // UpdateButton
            // 
            this.UpdateButton.Anchor = System.Windows.Forms.AnchorStyles.Top;
            this.UpdateButton.Enabled = false;
            this.UpdateButton.Font = new System.Drawing.Font("Microsoft Sans Serif", 8F, System.Drawing.FontStyle.Regular, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.UpdateButton.Location = new System.Drawing.Point(405, 21);
            this.UpdateButton.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.UpdateButton.Name = "UpdateButton";
            this.UpdateButton.Size = new System.Drawing.Size(237, 58);
            this.UpdateButton.TabIndex = 1;
            this.UpdateButton.Text = "&Install selected updates";
            this.UpdateButton.UseVisualStyleBackColor = true;
            this.UpdateButton.Click += new System.EventHandler(this.UpdateButton_Click);
            // 
            // UpdateList
            // 
            this.UpdateList.CheckBoxes = true;
            this.UpdateList.Columns.AddRange(new System.Windows.Forms.ColumnHeader[] {
            this.ComponentTitle,
            this.ComponentDescription,
            this.ComponentDate,
            this.ComponentSize});
            this.UpdateList.FullRowSelect = true;
            this.UpdateList.GridLines = true;
            this.UpdateList.HeaderStyle = System.Windows.Forms.ColumnHeaderStyle.Nonclickable;
            this.UpdateList.HideSelection = false;
            this.UpdateList.Location = new System.Drawing.Point(8, 91);
            this.UpdateList.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.UpdateList.MultiSelect = false;
            this.UpdateList.Name = "UpdateList";
            this.UpdateList.Size = new System.Drawing.Size(876, 296);
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
            this.SettingsTab.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.SettingsTab.Name = "SettingsTab";
            this.SettingsTab.Padding = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.SettingsTab.Size = new System.Drawing.Size(908, 407);
            this.SettingsTab.TabIndex = 2;
            this.SettingsTab.Text = "Settings";
            this.SettingsTab.UseVisualStyleBackColor = true;
            // 
            // SaveButton
            // 
            this.SaveButton.Location = new System.Drawing.Point(74, 271);
            this.SaveButton.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.SaveButton.Name = "SaveButton";
            this.SaveButton.Size = new System.Drawing.Size(240, 40);
            this.SaveButton.TabIndex = 8;
            this.SaveButton.Text = "Save and &restart";
            this.SaveButton.UseVisualStyleBackColor = true;
            this.SaveButton.Click += new System.EventHandler(this.SaveButton_Click);
            // 
            // BrowseButton
            // 
            this.BrowseButton.Location = new System.Drawing.Point(699, 66);
            this.BrowseButton.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.BrowseButton.Name = "BrowseButton";
            this.BrowseButton.Size = new System.Drawing.Size(112, 34);
            this.BrowseButton.TabIndex = 2;
            this.BrowseButton.Text = "&Browse";
            this.BrowseButton.UseVisualStyleBackColor = true;
            this.BrowseButton.Click += new System.EventHandler(this.BrowseButton_Click);
            // 
            // LocationBox
            // 
            this.LocationBox.Location = new System.Drawing.Point(240, 68);
            this.LocationBox.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.LocationBox.Name = "LocationBox";
            this.LocationBox.Size = new System.Drawing.Size(448, 26);
            this.LocationBox.TabIndex = 1;
            // 
            // LocationLabel
            // 
            this.LocationLabel.AutoSize = true;
            this.LocationLabel.Location = new System.Drawing.Point(78, 72);
            this.LocationLabel.Margin = new System.Windows.Forms.Padding(4, 0, 4, 0);
            this.LocationLabel.Name = "LocationLabel";
            this.LocationLabel.Size = new System.Drawing.Size(152, 20);
            this.LocationLabel.TabIndex = 0;
            this.LocationLabel.Text = "Flashpoint Location:";
            // 
            // CheckFilesButton
            // 
            this.CheckFilesButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.CheckFilesButton.Location = new System.Drawing.Point(330, 337);
            this.CheckFilesButton.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.CheckFilesButton.Name = "CheckFilesButton";
            this.CheckFilesButton.Size = new System.Drawing.Size(240, 40);
            this.CheckFilesButton.TabIndex = 9;
            this.CheckFilesButton.Text = "Check for &missing files";
            this.CheckFilesButton.UseVisualStyleBackColor = true;
            this.CheckFilesButton.Click += new System.EventHandler(this.CheckFilesButton_Click);
            // 
            // StableRepo
            // 
            this.StableRepo.AutoSize = true;
            this.StableRepo.Location = new System.Drawing.Point(304, 188);
            this.StableRepo.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.StableRepo.Name = "StableRepo";
            this.StableRepo.Size = new System.Drawing.Size(80, 24);
            this.StableRepo.TabIndex = 5;
            this.StableRepo.TabStop = true;
            this.StableRepo.Text = "&Stable";
            this.StableRepo.UseVisualStyleBackColor = true;
            this.StableRepo.CheckedChanged += new System.EventHandler(this.ComponentRepo_CheckedChanged);
            // 
            // RepositoryBox
            // 
            this.RepositoryBox.Location = new System.Drawing.Point(240, 148);
            this.RepositoryBox.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.RepositoryBox.Name = "RepositoryBox";
            this.RepositoryBox.Size = new System.Drawing.Size(448, 26);
            this.RepositoryBox.TabIndex = 4;
            // 
            // DevRepo
            // 
            this.DevRepo.AutoSize = true;
            this.DevRepo.Location = new System.Drawing.Point(396, 188);
            this.DevRepo.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.DevRepo.Name = "DevRepo";
            this.DevRepo.Size = new System.Drawing.Size(128, 24);
            this.DevRepo.TabIndex = 6;
            this.DevRepo.TabStop = true;
            this.DevRepo.Text = "&Development";
            this.DevRepo.UseVisualStyleBackColor = true;
            this.DevRepo.CheckedChanged += new System.EventHandler(this.ComponentRepo_CheckedChanged);
            // 
            // RepositoryLabel
            // 
            this.RepositoryLabel.AutoSize = true;
            this.RepositoryLabel.Location = new System.Drawing.Point(78, 152);
            this.RepositoryLabel.Margin = new System.Windows.Forms.Padding(4, 0, 4, 0);
            this.RepositoryLabel.Name = "RepositoryLabel";
            this.RepositoryLabel.Size = new System.Drawing.Size(151, 20);
            this.RepositoryLabel.TabIndex = 3;
            this.RepositoryLabel.Text = "Component Source:";
            // 
            // CustomRepo
            // 
            this.CustomRepo.AutoSize = true;
            this.CustomRepo.Location = new System.Drawing.Point(537, 188);
            this.CustomRepo.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.CustomRepo.Name = "CustomRepo";
            this.CustomRepo.Size = new System.Drawing.Size(89, 24);
            this.CustomRepo.TabIndex = 7;
            this.CustomRepo.TabStop = true;
            this.CustomRepo.Text = "&Custom";
            this.CustomRepo.UseVisualStyleBackColor = true;
            this.CustomRepo.CheckedChanged += new System.EventHandler(this.ComponentRepo_CheckedChanged);
            // 
            // UninstallButton
            // 
            this.UninstallButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.UninstallButton.Location = new System.Drawing.Point(580, 337);
            this.UninstallButton.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
            this.UninstallButton.Name = "UninstallButton";
            this.UninstallButton.Size = new System.Drawing.Size(240, 40);
            this.UninstallButton.TabIndex = 10;
            this.UninstallButton.Text = "&Uninstall Flashpoint";
            this.UninstallButton.UseVisualStyleBackColor = true;
            this.UninstallButton.Click += new System.EventHandler(this.UninstallButton_Click);
            // 
            // OfflineIndicator
            // 
            this.OfflineIndicator.AutoSize = true;
            this.OfflineIndicator.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Bold, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.OfflineIndicator.Location = new System.Drawing.Point(795, 25);
            this.OfflineIndicator.Margin = new System.Windows.Forms.Padding(4, 0, 4, 0);
            this.OfflineIndicator.Name = "OfflineIndicator";
            this.OfflineIndicator.Size = new System.Drawing.Size(116, 20);
            this.OfflineIndicator.TabIndex = 1;
            this.OfflineIndicator.Text = "Offline Mode";
            this.OfflineIndicator.Visible = false;
            // 
            // Main
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(9F, 20F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(949, 464);
            this.Controls.Add(this.OfflineIndicator);
            this.Controls.Add(this.TabControl);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.Icon = ((System.Drawing.Icon)(resources.GetObject("$this.Icon")));
            this.Margin = new System.Windows.Forms.Padding(4, 5, 4, 5);
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
            this.UpdateTab.PerformLayout();
            this.groupBox1.ResumeLayout(false);
            this.groupBox1.PerformLayout();
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
        private System.Windows.Forms.GroupBox groupBox1;
        public System.Windows.Forms.Label lblTotalUpdates;
        public System.Windows.Forms.Label lblTotalUpdatesSize;
        private System.Windows.Forms.CheckBox chkUncheckAll;
    }
}

