using System.Diagnostics;
using Microsoft.WindowsAPICodePack.Dialogs;

namespace FlashpointInstaller
{
    public partial class Main : Form
    {
        public Main() => InitializeComponent();

        private void Main_Load(object sender, EventArgs e)
        {
            FolderTextBox.AutoSize = false;
            FolderTextBox.Height = 21;

            FolderTextBox.Text = Environment.GetFolderPath(Environment.SpecialFolder.Desktop);
        }

        private void Link_LinkClicked(object sender, LinkLabelLinkClickedEventArgs e)
        {
            Process.Start(new ProcessStartInfo("https://bluemaxima.org/flashpoint") { UseShellExecute = true });
        }

        private void FolderButton_Click(object sender, EventArgs e)
        {
            CommonOpenFileDialog PathDialog = new() { IsFolderPicker = true };

            if (PathDialog.ShowDialog() == CommonFileDialogResult.Ok)
            {
                SetPath(PathDialog.FileName, true);
            }
        }

        private bool SetPath(string path, bool updateText)
        {
            string errorPath = "";

            if (path.StartsWith(Environment.GetFolderPath(Environment.SpecialFolder.ProgramFiles))
             || path.StartsWith(Environment.GetFolderPath(Environment.SpecialFolder.ProgramFilesX86)))
            {
                errorPath = "Program Files";
            }
            else if (path.StartsWith(Path.GetTempPath().Remove(Path.GetTempPath().Length - 1)))
            {
                errorPath = "Temporary Files";
            }
            else if (path.StartsWith(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile) + @"\OneDrive"))
            {
                errorPath = "OneDrive";
            }
            else
            {
                if (updateText)
                {
                    FolderTextBox.Text = path;
                }
                return true;
            }

            MessageBox.Show(
                $"Flashpoint cannot be installed to the {errorPath} directory! Choose a different folder.",
                "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
            );
            return false;
        }

        private void Install_Click(object sender, EventArgs e)
        {
            if (!SetPath(FolderTextBox.Text, false))
            {
                return;
            }

            Install InstallWindow = new();
            InstallWindow.ShowDialog();
        }
    }
}