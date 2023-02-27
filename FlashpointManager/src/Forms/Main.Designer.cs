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
            this.TabControl = new System.Windows.Forms.TabControl();
            this.ManageTab = new System.Windows.Forms.TabPage();
            this.Message = new System.Windows.Forms.Label();
            this.Message2 = new System.Windows.Forms.Label();
            this.ChangeButton = new System.Windows.Forms.Button();
            this.DescriptionBox = new System.Windows.Forms.GroupBox();
            this.Description = new System.Windows.Forms.Label();
            this.UpdateTab = new System.Windows.Forms.TabPage();
            this.UpdateButton = new System.Windows.Forms.Button();
            this.UpdateList = new System.Windows.Forms.ListView();
            this.ComponentTitle = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.ComponentDescription = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.ComponentUpdated = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.ComponentSize = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.RemoveTab = new System.Windows.Forms.TabPage();
            this.RemoveDescription = new System.Windows.Forms.Label();
            this.RemoveButton = new System.Windows.Forms.Button();
            this.ComponentList = new RikTheVeggie.TriStateTreeView();
            this.TabControl.SuspendLayout();
            this.ManageTab.SuspendLayout();
            this.DescriptionBox.SuspendLayout();
            this.UpdateTab.SuspendLayout();
            this.RemoveTab.SuspendLayout();
            this.SuspendLayout();
            // 
            // TabControl
            // 
            this.TabControl.Controls.Add(this.ManageTab);
            this.TabControl.Controls.Add(this.UpdateTab);
            this.TabControl.Controls.Add(this.RemoveTab);
            this.TabControl.ItemSize = new System.Drawing.Size(150, 20);
            this.TabControl.Location = new System.Drawing.Point(13, 10);
            this.TabControl.Name = "TabControl";
            this.TabControl.SelectedIndex = 0;
            this.TabControl.Size = new System.Drawing.Size(600, 240);
            this.TabControl.SizeMode = System.Windows.Forms.TabSizeMode.Fixed;
            this.TabControl.TabIndex = 5;
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
            this.Message.TabIndex = 20;
            this.Message.Text = "Choose components to add or remove from your Flashpoint copy.";
            // 
            // Message2
            // 
            this.Message2.Location = new System.Drawing.Point(410, 44);
            this.Message2.Name = "Message2";
            this.Message2.Size = new System.Drawing.Size(171, 30);
            this.Message2.TabIndex = 21;
            this.Message2.Text = "Click on a component or category to learn more about it.";
            // 
            // ChangeButton
            // 
            this.ChangeButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.ChangeButton.Location = new System.Drawing.Point(8, 176);
            this.ChangeButton.Name = "ChangeButton";
            this.ChangeButton.Size = new System.Drawing.Size(574, 26);
            this.ChangeButton.TabIndex = 16;
            this.ChangeButton.Text = "Apply changes";
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
            this.DescriptionBox.TabIndex = 14;
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
            this.UpdateButton.TabIndex = 17;
            this.UpdateButton.Text = "Install updates";
            this.UpdateButton.UseVisualStyleBackColor = true;
            this.UpdateButton.Click += new System.EventHandler(this.UpdateButton_Click);
            // 
            // UpdateList
            // 
            this.UpdateList.Columns.AddRange(new System.Windows.Forms.ColumnHeader[] {
            this.ComponentTitle,
            this.ComponentDescription,
            this.ComponentUpdated,
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
            this.UpdateList.TabIndex = 2;
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
            this.ComponentDescription.Width = 272;
            // 
            // ComponentUpdated
            // 
            this.ComponentUpdated.Text = "Last Updated";
            this.ComponentUpdated.Width = 90;
            // 
            // ComponentSize
            // 
            this.ComponentSize.Text = "Size";
            this.ComponentSize.Width = 65;
            // 
            // RemoveTab
            // 
            this.RemoveTab.Controls.Add(this.RemoveDescription);
            this.RemoveTab.Controls.Add(this.RemoveButton);
            this.RemoveTab.Location = new System.Drawing.Point(4, 24);
            this.RemoveTab.Name = "RemoveTab";
            this.RemoveTab.Padding = new System.Windows.Forms.Padding(3);
            this.RemoveTab.Size = new System.Drawing.Size(592, 212);
            this.RemoveTab.TabIndex = 2;
            this.RemoveTab.Text = "Uninstall Flashpoint";
            this.RemoveTab.UseVisualStyleBackColor = true;
            // 
            // RemoveDescription
            // 
            this.RemoveDescription.Location = new System.Drawing.Point(8, 50);
            this.RemoveDescription.Name = "RemoveDescription";
            this.RemoveDescription.Size = new System.Drawing.Size(574, 40);
            this.RemoveDescription.TabIndex = 20;
            this.RemoveDescription.Text = resources.GetString("RemoveDescription.Text");
            this.RemoveDescription.TextAlign = System.Drawing.ContentAlignment.MiddleCenter;
            // 
            // RemoveButton
            // 
            this.RemoveButton.Anchor = System.Windows.Forms.AnchorStyles.Bottom;
            this.RemoveButton.Location = new System.Drawing.Point(195, 113);
            this.RemoveButton.Name = "RemoveButton";
            this.RemoveButton.Size = new System.Drawing.Size(200, 26);
            this.RemoveButton.TabIndex = 17;
            this.RemoveButton.Text = "Uninstall Flashpoint";
            this.RemoveButton.UseVisualStyleBackColor = true;
            this.RemoveButton.Click += new System.EventHandler(this.RemoveButton_Click);
            // 
            // ComponentList
            // 
            this.ComponentList.Indent = 20;
            this.ComponentList.Location = new System.Drawing.Point(9, 12);
            this.ComponentList.Name = "ComponentList";
            this.ComponentList.Size = new System.Drawing.Size(391, 152);
            this.ComponentList.TabIndex = 13;
            this.ComponentList.TriStateStyleProperty = RikTheVeggie.TriStateTreeView.TriStateStyles.Installer;
            this.ComponentList.BeforeCheck += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeCheck);
            this.ComponentList.AfterCheck += new System.Windows.Forms.TreeViewEventHandler(this.ComponentList_AfterCheck);
            this.ComponentList.BeforeSelect += new System.Windows.Forms.TreeViewCancelEventHandler(this.ComponentList_BeforeSelect);
            // 
            // Main
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(624, 261);
            this.Controls.Add(this.TabControl);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.Icon = ((System.Drawing.Icon)(resources.GetObject("$this.Icon")));
            this.MaximizeBox = false;
            this.Name = "Main";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "Flashpoint Manager";
            this.Load += new System.EventHandler(this.Main_Load);
            this.TabControl.ResumeLayout(false);
            this.ManageTab.ResumeLayout(false);
            this.DescriptionBox.ResumeLayout(false);
            this.UpdateTab.ResumeLayout(false);
            this.RemoveTab.ResumeLayout(false);
            this.ResumeLayout(false);

        }

        #endregion
        private System.Windows.Forms.TabControl TabControl;
        private System.Windows.Forms.TabPage ManageTab;
        private System.Windows.Forms.TabPage RemoveTab;
        public RikTheVeggie.TriStateTreeView ComponentList;
        private System.Windows.Forms.GroupBox DescriptionBox;
        private System.Windows.Forms.Label Description;
        private System.Windows.Forms.Label Message;
        public System.Windows.Forms.Button ChangeButton;
        private System.Windows.Forms.Label RemoveDescription;
        public System.Windows.Forms.Button RemoveButton;
        public System.Windows.Forms.Label Message2;
        private System.Windows.Forms.TabPage UpdateTab;
        private System.Windows.Forms.ColumnHeader ComponentTitle;
        private System.Windows.Forms.ColumnHeader ComponentDescription;
        private System.Windows.Forms.ColumnHeader ComponentUpdated;
        private System.Windows.Forms.ColumnHeader ComponentSize;
        public System.Windows.Forms.Button UpdateButton;
        public System.Windows.Forms.ListView UpdateList;
    }
}

