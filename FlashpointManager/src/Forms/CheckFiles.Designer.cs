namespace FlashpointManager.src.Forms
{
    partial class CheckFiles
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
            System.ComponentModel.ComponentResourceManager resources = new System.ComponentModel.ComponentResourceManager(typeof(CheckFiles));
            this.FileList = new System.Windows.Forms.ListView();
            this.FileComponent = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.FileLocation = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.RedownloadButton = new System.Windows.Forms.Button();
            this.CancelButton = new System.Windows.Forms.Button();
            this.FileMessage = new System.Windows.Forms.LinkLabel();
            this.FileListLoading = new System.Windows.Forms.Label();
            this.SuspendLayout();
            // 
            // FileList
            // 
            this.FileList.Columns.AddRange(new System.Windows.Forms.ColumnHeader[] {
            this.FileComponent,
            this.FileLocation});
            this.FileList.Cursor = System.Windows.Forms.Cursors.Default;
            this.FileList.FullRowSelect = true;
            this.FileList.GridLines = true;
            this.FileList.HeaderStyle = System.Windows.Forms.ColumnHeaderStyle.Nonclickable;
            this.FileList.HideSelection = false;
            this.FileList.Location = new System.Drawing.Point(12, 53);
            this.FileList.MultiSelect = false;
            this.FileList.Name = "FileList";
            this.FileList.Size = new System.Drawing.Size(460, 160);
            this.FileList.Sorting = System.Windows.Forms.SortOrder.Ascending;
            this.FileList.TabIndex = 1;
            this.FileList.UseCompatibleStateImageBehavior = false;
            this.FileList.View = System.Windows.Forms.View.Details;
            this.FileList.Visible = false;
            // 
            // FileComponent
            // 
            this.FileComponent.Text = "Component";
            this.FileComponent.Width = 120;
            // 
            // FileLocation
            // 
            this.FileLocation.Text = "File Location";
            this.FileLocation.Width = 315;
            // 
            // RedownloadButton
            // 
            this.RedownloadButton.Enabled = false;
            this.RedownloadButton.Location = new System.Drawing.Point(11, 224);
            this.RedownloadButton.Name = "RedownloadButton";
            this.RedownloadButton.Size = new System.Drawing.Size(356, 26);
            this.RedownloadButton.TabIndex = 2;
            this.RedownloadButton.Text = "&Redownload components";
            this.RedownloadButton.UseVisualStyleBackColor = true;
            this.RedownloadButton.Click += new System.EventHandler(this.RedownloadButton_Click);
            // 
            // CancelButton
            // 
            this.CancelButton.Location = new System.Drawing.Point(373, 224);
            this.CancelButton.Name = "CancelButton";
            this.CancelButton.Size = new System.Drawing.Size(100, 26);
            this.CancelButton.TabIndex = 3;
            this.CancelButton.Text = "&Cancel";
            this.CancelButton.UseVisualStyleBackColor = true;
            this.CancelButton.Click += new System.EventHandler(this.CancelButton_Click);
            // 
            // FileMessage
            // 
            this.FileMessage.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Regular, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.FileMessage.LinkArea = new System.Windows.Forms.LinkArea(143, 14);
            this.FileMessage.Location = new System.Drawing.Point(12, 9);
            this.FileMessage.Name = "FileMessage";
            this.FileMessage.Size = new System.Drawing.Size(470, 40);
            this.FileMessage.TabIndex = 0;
            this.FileMessage.TabStop = true;
            this.FileMessage.Text = resources.GetString("FileMessage.Text");
            this.FileMessage.UseCompatibleTextRendering = true;
            this.FileMessage.LinkClicked += new System.Windows.Forms.LinkLabelLinkClickedEventHandler(this.FileMessage_LinkClicked);
            // 
            // FileListLoading
            // 
            this.FileListLoading.BackColor = System.Drawing.SystemColors.Control;
            this.FileListLoading.Location = new System.Drawing.Point(12, 117);
            this.FileListLoading.Name = "FileListLoading";
            this.FileListLoading.Size = new System.Drawing.Size(460, 23);
            this.FileListLoading.TabIndex = 6;
            this.FileListLoading.Text = "Searching for missing files...";
            this.FileListLoading.TextAlign = System.Drawing.ContentAlignment.MiddleCenter;
            // 
            // CheckFiles
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(484, 261);
            this.Controls.Add(this.FileListLoading);
            this.Controls.Add(this.FileMessage);
            this.Controls.Add(this.CancelButton);
            this.Controls.Add(this.RedownloadButton);
            this.Controls.Add(this.FileList);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedDialog;
            this.MaximizeBox = false;
            this.MinimizeBox = false;
            this.Name = "CheckFiles";
            this.ShowIcon = false;
            this.ShowInTaskbar = false;
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterParent;
            this.Text = "Missing File Checker";
            this.Load += new System.EventHandler(this.CheckFiles_Load);
            this.ResumeLayout(false);

        }

        #endregion

        private System.Windows.Forms.ListView FileList;
        private System.Windows.Forms.ColumnHeader FileComponent;
        private System.Windows.Forms.ColumnHeader FileLocation;
        private System.Windows.Forms.Button RedownloadButton;
        private System.Windows.Forms.Button CancelButton;
        private System.Windows.Forms.LinkLabel FileMessage;
        private System.Windows.Forms.Label FileListLoading;
    }
}