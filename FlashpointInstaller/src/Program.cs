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
            
            Application.Run(new Main());
        }
    }
}
