namespace FlashpointInstaller
{
    partial class FinishOperation
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
            this.Message = new System.Windows.Forms.Label();
            this.RunOnClose = new System.Windows.Forms.CheckBox();
            this.CloseButton = new System.Windows.Forms.Button();
            this.SuspendLayout();
            // 
            // Message
            // 
            this.Message.Location = new System.Drawing.Point(12, 10);
            this.Message.Name = "Message";
            this.Message.Size = new System.Drawing.Size(280, 20);
            this.Message.TabIndex = 0;
            this.Message.Text = "Flashpoint has been successfully downloaded.";
            this.Message.TextAlign = System.Drawing.ContentAlignment.MiddleCenter;
            // 
            // RunOnClose
            // 
            this.RunOnClose.AutoSize = true;
            this.RunOnClose.Checked = true;
            this.RunOnClose.CheckState = System.Windows.Forms.CheckState.Checked;
            this.RunOnClose.Location = new System.Drawing.Point(12, 45);
            this.RunOnClose.Name = "RunOnClose";
            this.RunOnClose.Size = new System.Drawing.Size(113, 17);
            this.RunOnClose.TabIndex = 1;
            this.RunOnClose.Text = "Launch Flashpoint";
            this.RunOnClose.UseVisualStyleBackColor = true;
            // 
            // CloseButton
            // 
            this.CloseButton.Location = new System.Drawing.Point(218, 42);
            this.CloseButton.Name = "CloseButton";
            this.CloseButton.Size = new System.Drawing.Size(75, 22);
            this.CloseButton.TabIndex = 2;
            this.CloseButton.Text = "Close";
            this.CloseButton.UseVisualStyleBackColor = true;
            this.CloseButton.Click += new System.EventHandler(this.FinishDownload_Exit);
            // 
            // FinishOperation
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(304, 73);
            this.Controls.Add(this.CloseButton);
            this.Controls.Add(this.RunOnClose);
            this.Controls.Add(this.Message);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.MaximizeBox = false;
            this.MinimizeBox = false;
            this.Name = "FinishOperation";
            this.ShowIcon = false;
            this.ShowInTaskbar = false;
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "Download Complete";
            this.TopMost = true;
            this.FormClosing += new System.Windows.Forms.FormClosingEventHandler(this.FinishDownload_Exit);
            this.Load += new System.EventHandler(this.FinishOperation_Load);
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private System.Windows.Forms.Label Message;
        private System.Windows.Forms.CheckBox RunOnClose;
        private System.Windows.Forms.Button CloseButton;
    }
}