namespace FlashpointInstaller
{
    partial class Operate
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
            this.ProgressMeasure = new System.Windows.Forms.ProgressBar();
            this.ProgressLabel = new System.Windows.Forms.Label();
            this.SuspendLayout();
            // 
            // ProgressMeasure
            // 
            this.ProgressMeasure.Location = new System.Drawing.Point(12, 10);
            this.ProgressMeasure.Maximum = 1000;
            this.ProgressMeasure.Name = "ProgressMeasure";
            this.ProgressMeasure.Size = new System.Drawing.Size(472, 23);
            this.ProgressMeasure.TabIndex = 1;
            // 
            // ProgressLabel
            // 
            this.ProgressLabel.AutoSize = true;
            this.ProgressLabel.Location = new System.Drawing.Point(12, 46);
            this.ProgressLabel.Name = "ProgressLabel";
            this.ProgressLabel.Size = new System.Drawing.Size(61, 13);
            this.ProgressLabel.TabIndex = 2;
            this.ProgressLabel.Text = "Preparing...";
            // 
            // Operate
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(496, 73);
            this.ControlBox = false;
            this.Controls.Add(this.ProgressLabel);
            this.Controls.Add(this.ProgressMeasure);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedDialog;
            this.MaximizeBox = false;
            this.MinimizeBox = false;
            this.Name = "Operate";
            this.ShowInTaskbar = false;
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "Modifying Flashpoint...";
            this.Load += new System.EventHandler(this.Operation_Load);
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion
        private System.Windows.Forms.ProgressBar ProgressMeasure;
        private System.Windows.Forms.Label ProgressLabel;
    }
}