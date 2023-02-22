using System;
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

            if (File.Exists(FPM.ConfigFile))
            {
                string[] config = File.ReadAllLines(FPM.ConfigFile);
                if (config.Length > 1) FPM.ListURL = config[1];
            }

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

            if (args.Length > 0)
            {
                switch (args[0])
                {
                    case "/manage":
                        FPM.StartupMode = 1;
                        break;
                    case "/update":
                        FPM.StartupMode = 2;
                        break;
                }
            }
            
            if (FPM.StartupMode != 2)
            {
                Application.Run(new Main());
            }
            else
            {
                Application.Run(new UpdateCheck());
            }
        }
    }
}
