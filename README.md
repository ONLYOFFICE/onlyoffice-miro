# ONLYOFFICE app for Miro

This app allows users to create and edit office documents from [Miro](https://miro.com/) using ONLYOFFICE Docs.

## App installation

Visit the [Miro Marketplace](https://miro.com/marketplace/) to install the ONLYOFFICE app. Once installed, the app will appear in the Miro sidebar for easy access.

## App configuration

After installation, follow these steps to configure the app.

Upon the first launch, you'll be directed to the Settings Page. Later, you can visit the Settings page anytime by clicking the gear icon within the ONLYOFFICE app.

Enter the required details for ONLYOFFICE Docs:

* Document Server Address
* [JWT Secret](https://helpcenter.onlyoffice.com/docs/installation/docs-configure-jwt.aspx)
* JWT Header

You can also connect to the public test server of ONLYOFFICE Docs for one month by checking the corresponding box.

## App usage

### Creating files

Open the ONLYOFFICE app and click the **Create new file** button in the File Manager window. Enter a file name and choose a format (Document, Spreadsheet, or Presentation), click the Create button. 

The file will be added to the Miro board at a random location. The ONLYOFFICE app will display a file list, including the newly created file.

### Editing existing files

You can collaboratively edit files created within the ONLYOFFICE app or those added to the Miro board. Click on a file line in the app to open it in the ONLYOFFICE editor in a new tab.

File options include:

* **Edit**: Opens the file in the ONLYOFFICE editor.
* **Navigate to**: Locates the file's position on the Miro board.
* **Download as PDF**: Downloads the file in PDF format.
* **Delete**: Removes the file from both the ONLYOFFICE app and the Miro board.

###  Important notes

Files uploaded directly to the Miro board will not automatically appear in the File Manager list of the ONLYOFFICE app. To refresh the list, click the corresponding icon in the File Manager window.

## Installing ONLYOFFICE Docs

To be able to work with office files within Miro, you will need an instance of [ONLYOFFICE Docs](https://www.onlyoffice.com/office-suite.aspx). You can install the self-hosted version of the editors or opt for ONLYOFFICE Docs Cloud which doesn't require downloading and installation.

**Self-hosted editors**

You can install [free Community version](https://www.onlyoffice.com/download-community.aspx#docs-community) of ONLYOFFICE Docs or scalable [Enterprise Edition](https://www.onlyoffice.com/download.aspx#docs-enterprise).

To install free Community version, use [Docker](https://github.com/onlyoffice/Docker-DocumentServer) (recommended) or follow [these instructions](https://helpcenter.onlyoffice.com/docs/installation/docs-community-install-ubuntu.aspx) for Debian, Ubuntu, or derivatives.

To install Enterprise Edition, follow the instructions [here](https://helpcenter.onlyoffice.com/docs/installation/enterprise).

**ONLYOFFICE Docs Cloud**

To get ONLYOFFICE Docs Cloud, get started [here](https://www.onlyoffice.com/docs-registration.aspx).

## Need help? Feedback & Support

In case of technical problems, the best way to get help is to submit your issues [here](https://github.com/ONLYOFFICE/onlyoffice-miro/issues). Alternatively, you can contact ONLYOFFICE team on [forum.onlyoffice.com](https://forum.onlyoffice.com/).