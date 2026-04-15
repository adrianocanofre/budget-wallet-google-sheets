function doPost(e) {
  try {
    Logger.log("==== START ====");

    if (!e || !e.postData || !e.postData.contents) {
      Logger.log("❌ No postData received");
      return ContentService.createTextOutput("No postData")
        .setMimeType(ContentService.MimeType.TEXT);
    }
    const SPREADSHEET_ID = "VALUE";
    const SHEET_NAME = "2026";
    const HEADER_ROW = 2;  // ← headers are in row 2
    const DATA_START_ROW = 3; // ← January is in row 3

    const sheet = SpreadsheetApp
      .openById(SPREADSHEET_ID)
      .getSheetByName(SHEET_NAME);

    const body = JSON.parse(e.postData.contents);
    const month = body.month;
    const data = body.data;

    // Read headers from row 2
    const headers = sheet.getRange(HEADER_ROW, 1, 1, sheet.getLastColumn()).getValues()[0];

    // Read months from column A, starting at row 2
    const monthValues = sheet.getRange(HEADER_ROW, 1, sheet.getLastRow(), 1).getValues().flat();

    const normalize = (s) => String(s).trim().toLowerCase();

    const rowIndex = monthValues.findIndex(m => normalize(m) === normalize(month));

    if (rowIndex === -1) {
      return ContentService.createTextOutput("Month not found: " + month)
        .setMimeType(ContentService.MimeType.TEXT);
    }

    const row = rowIndex + HEADER_ROW; // actual sheet row

    // Map from Go label → sheet header name
    const labelMap = {
        "50": "Rule 50",
        "30": "Rule 30",
        "20": "Rule 20",
        "Talho": "Talho",
        "Salary": "VWDS Salary",
        "Salary - PPR": "VWDS PPR",
        "Salary - Meal":  "Meal Card",
        "Transport": "Transport",
        "Trip": "Trip", 
        "Utilities": "Utilities",
        "Sports": "Sports",
        "Subscription": "Subscription",
        "Stock": "Stock",
        "Rent": "Rent",
        "Personal Care": "Personal Care",
        "Money Received": "Money Received",
        "Market": "Market",
        "Gadgets": "Gadgets",
        "Entertainment": "Entertainment",
        "Eating Out": "Eating Out",
        "Car - Parking": "Car - Parking",
        "Car - Fuel": "Car - Fuel",

    };

    Object.keys(data).forEach(label => {
      const headerName = labelMap[label] || label;
      const colIndex = headers.findIndex(h => normalize(h) === normalize(headerName));

      if (colIndex !== -1) {
        const value = parseFloat(data[label]) || 0;
        sheet.getRange(row, colIndex + 1).setValue(value);
        Logger.log(`✅ ${headerName} = ${value} → row ${row}, col ${colIndex + 1}`);
      } else {
        Logger.log(`⚠️ No header found for: ${headerName}`);
      }
    });

    return ContentService.createTextOutput("OK")
      .setMimeType(ContentService.MimeType.TEXT);

  } catch (err) {
    Logger.log("🔥 ERROR: " + err.toString());
    return ContentService.createTextOutput("ERROR: " + err.toString())
      .setMimeType(ContentService.MimeType.TEXT);
  }
}