using System;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using CyberSaaS.Terminal.Infrastructure.Api;

namespace CyberSaaS.Terminal.UI.Views;

public partial class StartSessionView : UserControl
{
    private readonly Action _cancelCallback;
    private readonly Func<TerminalStartSessionRequest, Task> _startCallback;
    private readonly TerminalCustomer _customer;

    public StartSessionView(
        Action cancelCallback,
        Func<TerminalStartSessionRequest, Task> startCallback,
        TerminalCustomer customer)
    {
        InitializeComponent();

        _cancelCallback = cancelCallback;
        _startCallback = startCallback;
        _customer = customer;

        CustomerNameTextBlock.Text =
            customer.FullName;

        CustomerPhoneTextBlock.Text =
            string.IsNullOrWhiteSpace(customer.Phone)
                ? "Phone: Not provided"
                : $"Phone: {customer.Phone}";

        CustomerIdTextBlock.Text =
            $"Customer type: {customer.CustomerType}";

        AdultSessionRadio.Checked += SessionTypeChanged;
        PrepaidSessionRadio.Checked += SessionTypeChanged;

        SessionTypeChanged(null!, null!);
    }

    private void SessionTypeChanged(
        object? sender,
        RoutedEventArgs? e)
    {
        if (AdultSessionRadio.IsChecked == true)
        {
            AdultSessionPanel.Visibility =
                Visibility.Visible;

            PrepaidSessionPanel.Visibility =
                Visibility.Collapsed;
        }
        else
        {
            AdultSessionPanel.Visibility =
                Visibility.Collapsed;

            PrepaidSessionPanel.Visibility =
                Visibility.Visible;
        }
    }

    private async void StartButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        string sessionType =
            AdultSessionRadio.IsChecked == true
                ? "pay_after"
                : "prepaid";

        string? prepaidAmount = null;

        if (sessionType == "prepaid")
        {
            if (!decimal.TryParse(
                    PrepaidAmountTextBox.Text.Trim(),
                    out decimal amount) ||
                amount <= 0)
            {
                StatusTextBlock.Text =
                    "Please enter a valid amount paid.";
                return;
            }

            prepaidAmount =
                amount.ToString("0.00");
        }

        StartButton.IsEnabled = false;

        StatusTextBlock.Text =
            "Starting session...";

        try
        {
            var request =
                new TerminalStartSessionRequest
                {
                    CustomerId =
                        _customer.Id,

                    SessionType =
                        sessionType,

                    PrepaidAmount =
                        prepaidAmount,

                    StartedAt =
                        DateTime.UtcNow,

                    ClientOperationId =
                        Guid.NewGuid().ToString()
                };

            await _startCallback(request);
        }
        catch (Exception ex)
        {
            StatusTextBlock.Text =
                $"Session could not be started.\n\n{ex.Message}";
        }
        finally
        {
            StartButton.IsEnabled = true;
        }
    }

    private void CancelButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        _cancelCallback();
    }
}