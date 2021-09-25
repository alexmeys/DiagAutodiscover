package main

import(
	"fmt"
	"net"
	"context"
	"time"
)

func main(){
	
	// Custom Public DNS resolver
	// ** Bug in Go 1.17 & Windows platform (github golang #33097)
	
	dns := &net.Resolver{
        PreferGo: true,
        Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
            d := net.Dialer{
                Timeout: time.Millisecond * time.Duration(6000),
            }
            return d.DialContext(ctx, network, "1.1.1.1:53")
	},
	}
	
	// Repeat the cycle
	for{
		// Getting the hostname
		host := ""
		fmt.Printf("\nEnter the domain name (domain.com): ")
		fmt.Scanln(&host)
		hosturl := "autodiscover."+host
		
		fmt.Print("\n---------------------* Start *----------------------\n\n")
		fmt.Printf("I Will test autodiscover for the domain: %s\n\n", host)
		
		// Check NS records
		nIP, err := dns.LookupNS(context.Background(), host)
		if err != nil{
			fmt.Println("\nThere are no Name Server records published, are you sure the domain exists? Try again, it might have been a typo.\n")
			continue
		} else {
			fmt.Print("------------------------------------------------------\n")
			fmt.Println("The servers who manage your DNS records:")
			fmt.Println("------------------------------------------------------")
			fmt.Println("If you need to make dns changes, these are the servers hosting the dns records:\n")
			for _, ns := range nIP{
				fmt.Printf("%s\n", ns.Host)
			}
			fmt.Print("\n")
		}
		
		// Check MX records 
		mREC, err := dns.LookupMX(context.Background(), host)
		if err != nil {
			fmt.Println("There is no mail records published. Why do you want autodiscover...\n")
			fmt.Println("Press the Enter Key to terminate the console screen!")
			fmt.Scanln()
			break
		} else {
			fmt.Print("----------------------------\n")
			fmt.Println("Inbound Mail Records:")
			fmt.Println("----------------------------")
			
			for _, mx := range mREC{
				fmt.Printf("%s - Priority: %v\n", mx.Host, mx.Pref)
			}
			fmt.Print("\n")
		}
		// Check A records & Srv Record
		aIP, err := dns.LookupHost(context.Background(), hosturl)
		fmt.Print("------------------------\n")
		fmt.Println("Certificate Check")
		fmt.Print("------------------------\n\n")
		if err != nil {
			// No A Records found
			fmt.Printf("NO A record found, %v.\n", err)
			fmt.Sprintf("This means you dont need the name %s in your certificate.\n", hosturl)
			fmt.Sprintf("In case you have it in your certificate, I advice you to create the A record %s \n", hosturl)
			fmt.Println("\nLet's see if there is a srv record, there shoud be one:")
			// Check SRV Records
			cname, sIP, err := dns.LookupSRV(context.Background(), "autodiscover", "tcp", host)
			if err != nil{
				fmt.Printf("NO SRV record found, %v.\n", err)
				fmt.Printf("It seems nothing is configured... May I advice you to create an srv record _autodiscover._tcp.%s and point it to your target mailserver (with a valid certificate name).\n", host)
				fmt.Print("It could be that you don't have any external configuration, only internally, but this would mean you know what you are doing and ... well why are you using this tool...\n")
			} else {
				fmt.Print(cname)
				fmt.Println("SRV record found, you don't need the autodiscover name in the actual certificate.\nIf the domain name is correct, this should work for autodiscover.\n\nDetails:")
				for _, srvip := range sIP{
					fmt.Printf("Target: %v, Port: %v, Priority: %d, Weight: %d\n\n", srvip.Target, srvip.Port, srvip.Priority, srvip.Weight)
					srvip.Target = srvip.Target[:len(srvip.Target)-1]
					fmt.Printf("Be Certain that the target name (%v) has a valid certificate. You should be able to browse and see: https://%v/owa without any warnings\n", srvip.Target, srvip.Target)
					fmt.Printf("Try logging in with a domain account to: https://%v/autodiscover/autodiscover.xml, if this gives issues, it might be an internal issue.\n", srvip.Target)
				}
			}
		} else {
			// A Records found
			fmt.Println("Certificate (A) Record Found\n")
			if len(aIP) > 1{
				fmt.Println("It seems you are using a Cname record to distribute your A records, most probably O365? Check in your O365 portal if everything is set-up correctly.")
			} else {
				// Check SRV records even with A record in place
				for _, ip := range aIP{
					_, sIP2, err := dns.LookupSRV(context.Background(), "autodiscover", "tcp", host)
					if err != nil{
						fmt.Printf("Is the name (%s) listed in the certificate, otherwise you will get autodiscover issues.\nIt points to this IP (WAN of the mailserver): %s\n", hosturl, ip)
						fmt.Printf("This url should work without any certificate warnings: https://%v/owa\n", hosturl)
						fmt.Printf("If that works, try logging in at https://%v/autodiscover/autodiscover.xml If this fails it might be an internal issue.\n", hosturl)
					} else {
						fmt.Println("You seem to have a SRV record AND an A record!")
						fmt.Printf("By design the A record (%s) will be preferred.\nIn case the name is not listed in the actual certificate (or alt cert names), please remove the A Record: %s and use the existing srv record.\n", hosturl, hosturl)
						fmt.Println("\nCheck that the srv record is valid. The record currently configured:")
						for _, srvip2 := range sIP2{
							fmt.Printf("Target: %v, Port: %v, Priority: %d, Weight: %d\n", srvip2.Target, srvip2.Port, srvip2.Priority, srvip2.Weight)
							srvip2.Target = srvip2.Target[:len(srvip2.Target)-1]
							fmt.Printf("\nFor this to work, the url: https://%v/owa should open without any warnings.\n", srvip2.Target)
							fmt.Printf("If that works, try https://%v/autodiscover/autodiscover.xml If this fails, it might be an internal issue.\n", srvip2.Target)
						}
					}
				}
			}
		}
		fmt.Print("\n---------------------* The End *----------------------\n")
		fmt.Println("If everything seems to be configured correctly, but there are still issues, check these:")
		fmt.Println("*\tIs there a local XML/Registry key set on the computer?")
		fmt.Println("*\tAre you using Office 365 Outlook, but with a local Exchange, take a look at the regkey: ExcludeO365Endpoint")
		fmt.Println("*\tIf it fails when adding the account, check for 2FA.")
		fmt.Print("\n------------------------------------------------------\n\n")
		fmt.Print("Want to run the wizard again (y/n)? ")
		
		var answ string
		fmt.Scanln(&answ)
		if answ != "y"{
			break
		}
		
		fmt.Print("\n******************************************************\n")
		fmt.Print("\n******************************************************\n")
	}
}
