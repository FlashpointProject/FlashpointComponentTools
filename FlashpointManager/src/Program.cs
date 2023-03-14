using System;
using System.Diagnostics;
using System.IO;
using System.Linq;
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
