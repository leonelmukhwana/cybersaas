using System;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;

namespace CyberSaaS.Terminal.UI.Views;

public partial class CustomerSelectionView : UserControl
{
    private readonly Func<string, string, Task> _lookupCustomerCallback;
    private readonly Action _cancelCallback;

    public CustomerSelectionView(
        Func<string, string, Task> lookupCustomerCallback,
        Action cancelCallback)
    {
        InitializeComponent();

        _lookupCustomerCallback =
            lookupCustomerCallback
            ?? throw new ArgumentNullException(
                nameof(lookupCustomerCallback));

        _cancelCallback =
            cancelCallback
            ?? throw new ArgumentNullException(
                nameof(cancelCallback));
    }

    private async void LookupButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        string searchTerm =
            SearchTextBox.Text.Trim();

        if (string.IsNullOrWhiteSpace(searchTerm))
        {
            StatusTextBlock.Text =
                "Please enter a customer search value.";

            SearchTextBox.Focus();

            return;
        }

        string searchType =
            AdultRadioButton.IsChecked == true
                ? "adult"
                : "child";

        try
        {
            StatusTextBlock.Text =
                "Searching for customer...";

            await _lookupCustomerCallback(
                searchType,
                searchTerm);
        }
        catch (Exception ex)
        {
            StatusTextBlock.Text =
                ex.Message;
        }
    }

    private void CancelButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        _cancelCallback();
    }

    private void CustomerType_Checked(
        object sender,
        RoutedEventArgs e)
    {
        if (SearchTextBox == null ||
            SearchLabel == null ||
            SearchHelpTextBlock == null)
        {
            return;
        }

        if (AdultRadioButton.IsChecked == true)
        {
            SearchLabel.Text =
                "ID Number";

            SearchTextBox.Clear();

            SearchHelpTextBlock.Text =
                "Enter the adult customer's national ID number.";

            SearchTextBox.ToolTip =
                "Adult ID number";

            return;
        }

        SearchLabel.Text =
            "Child's Name";

        SearchTextBox.Clear();

        SearchHelpTextBlock.Text =
            "Enter the child's full name.";

        SearchTextBox.ToolTip =
            "Child's full name";
    }
}