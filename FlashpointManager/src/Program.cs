using System;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Net;
using System.Windows.Forms;
using System.Xml;

using FlashpointManager.Common;
using Microsoft.WindowsAPICodePack.Dialogs;

namespace FlashpointManager
{
    internal static class Program
    {
        /// <summary>
        /// The main entry point for the application.
        /// </summary>
        [STAThread]
        static void Main(string[] args)
        {
            // Run app from temp folder as long as /notemp argument isn't passed
            // This allows the executable to be deleted when updating its component or uninstalling Flashpoint
            if (!args.Any(v => v.ToLower() == "/notemp") && !Debugger.IsAttached)
            {
                string realPath = AppDomain.CurrentDomain.BaseDirectory;
                string realFile = AppDomain.CurrentDomain.FriendlyName;
                string tempPath = Path.GetTempPath();
                string tempFile = $"69McIKvK_{realFile}";

                if (realPath != tempPath && realFile != tempFile)
                {
                    File.Copy(realPath + realFile, tempPath + tempFile, true);
                    Process.Start(new ProcessStartInfo
                    {
                        FileName = tempPath + tempFile,
                        Arguments = string.Join(" ", args),
                        WorkingDirectory = realPath
                    });
                    Environment.Exit(0);
                }
                else if (tempPath.TrimEnd('\\') == Directory.GetCurrentDirectory())
                {
                    Environment.Exit(0);
                }
            }

            Application.EnableVisualStyles();
            Application.SetCompatibleTextRenderingDefault(false);

            ServicePointManager.SecurityProtocol = SecurityProtocolType.Tls12;

            // Load config, or create if it doesn't exist
            try
            {
                var configReader = File.ReadAllLines(FPM.ConfigFile);
                FPM.SourcePath = configReader[0];
                FPM.RepoXml    = configReader[1];
            }
            catch
            {
                using (var configWriter = File.CreateText(FPM.ConfigFile))
                {
                    configWriter.WriteLine(FPM.SourcePath);
                    configWriter.WriteLine(FPM.RepoXml);
                }
            }

            Stream listStream = null;

            // Attempt to download and parse component list
            try
            {
                listStream = new MemoryStream(new WebClient().DownloadData(FPM.RepoXml));
                FPM.XmlTree.Load(listStream);
            }
            catch (Exception e)
            {
                string failedAction = e is XmlException ? "parsed" : "downloaded";

                FPM.GenericError(
                    $"The component list could not be {failedAction}! An offline backup will be used instead.\n\n" +
                    "Verify that your internet connection is working and that your component source is not misconfigured."
                );

                FPM.OfflineMode = true;
            }

            string backupPath = Path.Combine(FPM.SourcePath, "Components", "components.bak");

            // Create backup if component list was successfully downloaded and parsed; otherwise, load from backup
            if (!FPM.OfflineMode)
            {
                try
                {
                    using (var fileStream = new FileStream(backupPath, FileMode.OpenOrCreate, FileAccess.Write))
                    {
                        fileStream.SetLength(0);
                        listStream.Position = 0;
                        listStream.CopyTo(fileStream);
                    }
                }
                catch { }
            }
            else
            {
                try
                {
                    listStream = new FileStream(backupPath, FileMode.Open, FileAccess.Read);
                    FPM.XmlTree.Load(listStream);
                }
                catch
                {
                    FPM.GenericError("Failed to load component list from offline backup!");
                    Environment.Exit(1);
                }
            }

            // Verify that the configured Flashpoint path is valid
            while (!FPM.VerifySourcePath(FPM.SourcePath))
            {
                FPM.GenericError(
                    "The Flashpoint directory specified in fpm.cfg is invalid!\n\n" + 
                    "Please choose a valid directory."
                );

                var pathDialog = new CommonOpenFileDialog() { IsFolderPicker = true };

                if (pathDialog.ShowDialog() == CommonFileDialogResult.Cancel)
                {
                    Environment.Exit(1);
                }

                FPM.SourcePath = pathDialog.FileName;

                if (!FPM.WriteConfig(FPM.SourcePath, FPM.RepoXml))
                {
                    Environment.Exit(1);
                }
            }

            if (args.Length > 0)
            {
                // Open update tab on startup if /update argument is passed
                if (args.Any(v => v.ToLower() == "/update"))
                {
                    FPM.OpenUpdateTab = true;
                }

                // Open launcher on close if /launcher argument is passed
                if (args.Any(v => v.ToLower() == "/launcher"))
                {
                    FPM.OpenLauncherOnClose = true;
                }
                
                // Automatically download components if /download argument is passed
                var argsList = args.ToList();
                int first = argsList.FindIndex(v => v.ToLower() == "/download");

                if (first > -1 && first < argsList.Count - 1)
                {
                    int last = argsList.FindIndex(first + 1, v => v.StartsWith("/"));
                    if (last == -1) last = argsList.Count;

                    FPM.AutoDownload.AddRange(argsList.Skip(first + 1).Take(last - first - 1));
                }
            }

            // Display the application window
            Application.Run(new Main() { Opacity = FPM.AutoDownload.Count == 0 ? 1 : 0 });
        }
    }
}
