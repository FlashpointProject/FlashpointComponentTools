﻿using System;
using System.Diagnostics;
using System.IO;
using System.Threading.Tasks;
using System.Windows.Forms;

using FlashpointManager.Common;
using FlashpointManager.src.Forms;
using Microsoft.WindowsAPICodePack.Dialogs;

namespace FlashpointManager
{
    public partial class Main : Form
    {
        private void BrowseButton_Click(object sender, EventArgs e)
        {
            var pathDialog = new CommonOpenFileDialog() { IsFolderPicker = true };

            if (pathDialog.ShowDialog() == CommonFileDialogResult.Ok)
            {
                if (FPM.VerifySourcePath(pathDialog.FileName))
                {
                    LocationBox.Text = pathDialog.FileName;
                }
                else
                {
                    FPM.GenericError("The specified Flashpoint directory is invalid!");
                }
            }
        }

        private void SaveButton_Click(object sender, EventArgs e)
        {
            if (!FPM.WriteConfig(LocationBox.Text, RepositoryBox.Text)) return;

            Application.Restart();
        }

        private void CheckFilesButton_Click(object sender, EventArgs e) => new CheckFiles().ShowDialog();

        private void ComponentRepo_CheckedChanged(object sender, EventArgs e)
        {
            var radio = sender as RadioButton;
            if (!radio.Checked) return;

            if (radio.Name == "StableRepo")
            {
                RepositoryBox.Text = FPM.RepoXmlTemplates.Stable;
            }
            else if (radio.Name == "DevRepo")
            {
                RepositoryBox.Text = FPM.RepoXmlTemplates.Development;
            }
            else
            {
                RepositoryBox.Text = FPM.RepoXml;
            }

            RepositoryBox.Enabled = radio.Name == "CustomRepo";
        }

        private async void UninstallButton_Click(object sender, EventArgs e)
        {
            var uninstallDialog = MessageBox.Show(
                "Understand that uninstalling Flashpoint will delete:\n\n" +
                "- Downloaded entries\n" +
                "- Save data (depending on the entry)\n" + 
                "- Playlists\n" +
                "- Favorites\n\n" +
                "This operation cannot be cancelled once started. Are you sure you want to continue?",
                "Warning", MessageBoxButtons.YesNo, MessageBoxIcon.Warning
            );

            if (uninstallDialog != DialogResult.Yes) return;

            TabControl.Enabled = false;

            await Task.Run(() => {
                foreach (string file in Directory.EnumerateFiles(FPM.SourcePath, "*", SearchOption.AllDirectories))
                {
                    try { FPM.DeleteFileAndDirectories(file); } catch { }
                }

                var shortcutPaths = new string[]
                {
                    Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.Programs),  "Flashpoint.lnk"),
                    Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.StartMenu), "Flashpoint.lnk"),
                    Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.Desktop),   "Flashpoint.lnk")
                };

                foreach (string path in shortcutPaths)
                {
                    try { File.Delete(path); } catch { }
                }
            });

            MessageBox.Show(
                "Flashpoint has been uninstalled from your system.",
                "Uninstallation Complete", MessageBoxButtons.OK, MessageBoxIcon.Information
            );

            new Process()
            {
                StartInfo = {
                FileName = "cmd.exe",
                Arguments = "/k timeout /nobreak /t 1 >nul & " +
                    $"rmdir /Q \"{Path.GetFullPath(Path.Combine(FPM.SourcePath, "Manager"))}\" & " +
                    $"rmdir /Q \"{Path.GetFullPath(FPM.SourcePath)}\" & exit",
                WorkingDirectory = Path.GetFullPath(Path.Combine(FPM.SourcePath, "..")),
                WindowStyle = ProcessWindowStyle.Hidden
            }
            }.Start();

            Close();
        }
    }
}
