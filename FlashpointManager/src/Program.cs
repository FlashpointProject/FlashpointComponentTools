using System;
using System.Diagnostics;
using System.IO;
using System.Threading.Tasks;
using System.Windows.Forms;
using System.Xml;

using Downloader;
using FlashpointInstaller.Common;

namespace FlashpointInstaller
{
    internal static class Program
    {
        /// <summary>
        /// The main entry point for the application.
        /// </summary>
        [STAThread]
        static void Main(string[] args)
        {
            // Run app from temp folder if it isn't already
            // This allows the executable to be deleted when updating the component or uninstalling Flashpoint
            if (!Debugger.IsAttached)
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

            // Load config, or create if it doesn't exist
            try
            {
                var configReader = File.ReadAllLines(FPM.ConfigFile);
                FPM.SourcePath = configReader[0];
                FPM.ListURL    = configReader[1];
            }
            catch
            {
                using (var configWriter = File.CreateText(FPM.ConfigFile))
                {
                    configWriter.WriteLine(FPM.SourcePath);
                    configWriter.WriteLine(FPM.ListURL);
                }
            }

            // Download and parse component list
            Stream listStream = null;
            Task.Run(async () => { listStream = await new DownloadService().DownloadFileTaskAsync(FPM.ListURL); }).Wait();

            if (listStream == null)
            {
                MessageBox.Show(
                    "The component list could not be downloaded! Do you have an internet connection?",
                    "Error", MessageBoxButtons.OK, MessageBoxIcon.Error
                );

                Environment.Exit(1);
            }

            listStream.Position = 0;

            FPM.XmlTree = new XmlDocument();
            FPM.XmlTree.Load(listStream);

            // Verify that the configured Flashpoint path is valid
            FPM.VerifySourcePath();

            if (args.Length > 0)
            {
                // Open update tab on startup if /update argument is passed
                if (args[0].ToLower() == "/update")
                {
                    FPM.OpenUpdateTab = true;
                }
                // Automatically download component if /download argument is passed
                else if (args.Length > 1 && args[0].ToLower() == "/download")
                {
                    FPM.AutoDownload = args[1].ToLower();
                }
            }

            // Display the application window
            Application.Run(new Main() { Opacity = FPM.AutoDownload.Length > 0 ? 0 : 1 });
        }
    }
}
