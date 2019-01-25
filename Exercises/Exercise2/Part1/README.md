# Mutex and Channel basics

### What is an atomic operation?
> *En atomisk operasjon er en operasjon som alltid vil bli utført uten at noen andre prosesser får mulitheten til å lese eller endre tilstanden som leses eller endres under operasjonen (https://goo.gl/wESWcE).*

### What is a semaphore?
> *En semafor er en variabel som brukes for å kontrollere tilgangen til en felles ressurs (https://goo.gl/ySMh1). Semaforen kan ses på som en integer, med tre forskjeller:*
>-  *Når du oppretter en semafor, så kan du initialisere den som et hvilket som helst heltall, men etter det har du bare muligheten til å inkrementere eller dekrementere. Du kan ikke lese semaforens nåverende verdi.*
>-  *Når en tråd dekrementerer semaforen, vil tråden blokkere seg selv dersom dersom resultatet er negativt, og fortsetter ikke før en annen tråd inkrementerer semaforen.*
>-  *Når en tråd inkrementerer semaforen, vil en av de ventende trådene (om noen) kunne fortsette.*

>*(https://goo.gl/wNgtsZ)*

### What is a mutex?
> *Mutex står for "mutual exclusion". En mutex er som en token, hvor en tråd må ha tokenen for å fortsette. Dette brukes for å for eksempel hindre at flere tråder bruker den samme variabelen samtidig: en tråd "tar" mutexen før den skal f.eks. endre variabelen, for så å "gi fra seg" mutexen når den er ferdig. Bare én tråd kan holde mutexen samtidig (https://goo.gl/wNgtsZ).*

### What is the difference between a mutex and a binary semaphore?
> *En binær semafor og en mutex er nokså like, men det er forskjeller i hvordan de brukes. En binær semafor kan brukes som en mutex, men en mutex har den begrensningen/egenskapen at bare tråden som har mutexen kan gi fra seg mutexen. Om du bruker en binær semafor, så kan tråd A ta semaforen, mens tråd B kan gi den vekk igjen (https://goo.gl/ySMh1, https://goo.gl/wb7vAK).*

### What is a critical section?
> *En critical section er en seksjon i koden hvor det er kritisk viktig å hindre concurrent tilgang (https://goo.gl/wNgtsZ, s. 19).*

### What is the difference between race conditions and data races?
> *Flere ønsker å bruker en delt ressurs samtidig. Én tråd kan da f.eks. overskrive arbeidet en annen tråd hadde gjort. Resultatet er da avhengig av når/i hvilken rekkefølge de ulike trådenes operasjoner kjøres. Om vi ikke har kontroll over denne rekkefølgen, så har vi en race condition. Timing/rekkefølge påvirker programmets korrekthet.*
> *Et data race oppstår når to instruksjoner fra ulike tråder aksesserer den samme minnelokasjonen, minst en av disse aksesseringene er for å skrive, og det er ingen synkronisering på plass som bestemmer rekkefølgen på disse aksesseringene.*
> *Om du har to enkle tråder som begge bruker mutex for å beskytte den kritiske variabelen i, og den ene tråden ønsker å sette i = 1 mens den andre tråder ønsker å sette i = 2, så har du ikke et data race -- mutexen hindrer at trådene får aksessere minnelokasjonen samtidig, og dermed er vi utenfor definisjonen av et data race. Men, det kan fremdeles oppstå en race condition, da vi ikke har noen mekanisme på plass som bestemmer hvilken av trådene som skal kjøre først. Dette kan påvirke programmets korrekthet (https://goo.gl/1x8jK4).*

### List some advantages of using message passing over lock-based synchronization primitives.
>- *Lettere å skalere (eksempel fra forelesning: internett)*
>- *Om du først oppnår korrekthet, så er optimal ytelse ikke så langt unna (https://goo.gl/zofJP3) (https://goo.gl/ETwz7V, side 17)* 

### List some advantages of using lock-based synchronization primitives over message passing.
>- *Lettere å oppnå korrekthet enn ved message passing (https://goo.gl/zofJP3)*
>- *Lettere å gradvis gjøre løsningen din bedre (https://goo.gl/ETwz7V, side 17)*
