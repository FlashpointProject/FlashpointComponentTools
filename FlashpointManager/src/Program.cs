using System;
using System.Collections.Generic;
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
            Application.EnableVisualStyles();
            Application.SetCompatibleTextRenderingDefault(false);

            var config = new List<string>() { Path.GetFullPath(Path.Combine(AppContext.BaseDirectory, "..")), FPM.ListURL };

            try
            {
                var configReader = File.ReadAllLines("fpm.cfg");
                config[0] = configReader[0];
                config[1] = configReader[1];
            }
            catch
            {
                using (var configWriter = File.CreateText("fpm.cfg"))
                {
                    configWriter.WriteLine(config[0]);
                    configWriter.WriteLine(config[1]);
                }
            }

            FPM.SourcePath = config[0];
            FPM.ListURL    = config[1];

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

            FPM.VerifySourcePath();

            if (args.Length > 0 && args[0].ToLower() == "/update") FPM.OpenUpdateTab = true;

            Application.Run(new Main());
        }
    }
}
