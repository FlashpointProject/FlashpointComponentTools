namespace FlashpointInstaller
{
    partial class UpdateCheck
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
            this.UpdateMessage = new System.Windows.Forms.Label();
            this.UpdateList = new System.Windows.Forms.ListView();
            this.ComponentTitle = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.ComponentDescription = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.ComponentSize = ((System.Windows.Forms.ColumnHeader)(new System.Windows.Forms.ColumnHeader()));
            this.UpdateButton = new System.Windows.Forms.Button();
            this.CloseButton = new System.Windows.Forms.Button();
            this.UpdateSizeLabel = new System.Windows.Forms.Label();
            this.UpdateSizeDisplay = new System.Windows.Forms.Label();
            this.SuspendLayout();
            // 
            // UpdateMessage
            // 
            this.UpdateMessage.Location = new System.Drawing.Point(12, 11);
            this.UpdateMessage.Name = "UpdateMessage";
            this.UpdateMessage.Size = new System.Drawing.Size(510, 16);
            this.UpdateMessage.TabIndex = 0;
            this.UpdateMessage.Text = "Available updates for installed components are displayed below.";
            // 
            // UpdateList
            // 
            this.UpdateList.BorderStyle = System.Windows.Forms.BorderStyle.FixedSingle;
            this.UpdateList.Columns.AddRange(new System.Windows.Forms.ColumnHeader[] {
            this.ComponentTitle,
            this.ComponentDescription,
            this.ComponentSize});
            this.UpdateList.FullRowSelect = true;
            this.UpdateList.GridLines = true;
            this.UpdateList.HeaderStyle = System.Windows.Forms.ColumnHeaderStyle.Nonclickable;
            this.UpdateList.HideSelection = false;
            this.UpdateList.Location = new System.Drawing.Point(12, 34);
            this.UpdateList.MultiSelect = false;
            this.UpdateList.Name = "UpdateList";
            this.UpdateList.Size = new System.Drawing.Size(510, 232);
            this.UpdateList.Sorting = System.Windows.Forms.SortOrder.Ascending;
            this.UpdateList.TabIndex = 1;
            this.UpdateList.UseCompatibleStateImageBehavior = false;
            this.UpdateList.View = System.Windows.Forms.View.Details;
            // 
            // ComponentTitle
            // 
            this.ComponentTitle.Text = "Title";
            this.ComponentTitle.Width = 122;
            // 
            // ComponentDescription
            // 
            this.ComponentDescription.Text = "Description";
            this.ComponentDescription.Width = 304;
            // 
            // ComponentSize
            // 
            this.ComponentSize.Text = "Size";
            // 
            // UpdateButton
            // 
            this.UpdateButton.Enabled = false;
            this.UpdateButton.Location = new System.Drawing.Point(152, 277);
            this.UpdateButton.Name = "UpdateButton";
            this.UpdateButton.Size = new System.Drawing.Size(288, 23);
            this.UpdateButton.TabIndex = 2;
            this.UpdateButton.Text = "Update Flashpoint";
            this.UpdateButton.UseVisualStyleBackColor = true;
            this.UpdateButton.Click += new System.EventHandler(this.UpdateButton_Click);
            // 
            // CloseButton
            // 
            this.CloseButton.Location = new System.Drawing.Point(448, 277);
            this.CloseButton.Name = "CloseButton";
            this.CloseButton.Size = new System.Drawing.Size(75, 23);
            this.CloseButton.TabIndex = 3;
            this.CloseButton.Text = "Close";
            this.CloseButton.UseVisualStyleBackColor = true;
            this.CloseButton.Click += new System.EventHandler(this.CloseButton_Click);
            // 
            // UpdateSizeLabel
            // 
            this.UpdateSizeLabel.AutoSize = true;
            this.UpdateSizeLabel.Font = new System.Drawing.Font("Microsoft Sans Serif", 8.25F, System.Drawing.FontStyle.Bold, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.UpdateSizeLabel.Location = new System.Drawing.Point(12, 282);
            this.UpdateSizeLabel.Margin = new System.Windows.Forms.Padding(0);
            this.UpdateSizeLabel.Name = "UpdateSizeLabel";
            this.UpdateSizeLabel.Size = new System.Drawing.Size(81, 13);
            this.UpdateSizeLabel.TabIndex = 4;
            this.UpdateSizeLabel.Text = "Size change:";
            // 
            // UpdateSizeDisplay
            // 
            this.UpdateSizeDisplay.AutoSize = true;
            this.UpdateSizeDisplay.Location = new System.Drawing.Point(93, 282);
            this.UpdateSizeDisplay.Margin = new System.Windows.Forms.Padding(0);
            this.UpdateSizeDisplay.Name = "UpdateSizeDisplay";
            this.UpdateSizeDisplay.Size = new System.Drawing.Size(38, 13);
            this.UpdateSizeDisplay.TabIndex = 5;
            this.UpdateSizeDisplay.Text = "0.0MB";
            // 
            // UpdateCheck
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(534, 311);
            this.Controls.Add(this.UpdateSizeDisplay);
            this.Controls.Add(this.UpdateSizeLabel);
            this.Controls.Add(this.CloseButton);
            this.Controls.Add(this.UpdateButton);
            this.Controls.Add(this.UpdateList);
            this.Controls.Add(this.UpdateMessage);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedDialog;
            this.MaximizeBox = false;
            this.MinimizeBox = false;
            this.Name = "UpdateCheck";
            this.ShowIcon = false;
            this.ShowInTaskbar = false;
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "Update Checker";
            this.Load += new System.EventHandler(this.UpdateCheck_Load);
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private System.Windows.Forms.Label UpdateMessage;
        private System.Windows.Forms.ListView UpdateList;
        private System.Windows.Forms.Button UpdateButton;
        private System.Windows.Forms.Button CloseButton;
        private System.Windows.Forms.ColumnHeader ComponentTitle;
        private System.Windows.Forms.ColumnHeader ComponentDescription;
        private System.Windows.Forms.ColumnHeader ComponentSize;
        private System.Windows.Forms.Label UpdateSizeLabel;
        private System.Windows.Forms.Label UpdateSizeDisplay;
    }
}