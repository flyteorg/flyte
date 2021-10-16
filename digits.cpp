/*  C++ Program to Find the Number of Digits in a number  */

#include<iostream>
using namespace std;

int main()
{
    int n,no,a=0;

    cout<<"Enter any positive integer :: ";
    cin>>n;

    no=n;

    while(no>0)
    {
        no=no/10;
        a++;
    }
    cout<<"\nNumber of Digits in a number [ "<<n<<" ] is :: "<<a<<"\n";

   return 0;
}
